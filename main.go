package main

import (
	"fmt"
	_"reflect"
	"time"
	"io/ioutil"
	"log"
	"net/http"
	"encoding/json"
	"github.com/gorilla/feeds"
)

// PULL DATA FROM API AND PARSE IT TO OBJECT
type vocaResponse struct {
	Items []vocaItems `json:"items"`
	TypeOfPull string
}
type vocaItems struct {
	ID int `json:"id",omitempty`
	CreateDate string `json:"createDate",omitempty`
	ArtistString string `json:"artistString",omitempty`
	Name string `json:"name",omitempty`
	DefaultNameLanguage string `json:"defaultNameLanguage"`
	Status bool `json:"status",omitempty`
	ReleaseDate albumsReleaseDate `json:"releaseDate",omitempty`
	PublishDate string `json:"publishDate",omitempty` //RETURNED ONLY IN SONGS
}
type albumsReleaseDate struct {
	Formatted string `json:"formatted",omitempty` //RETURNED ONLY IN ALBUMS
}

func pullLatestAlbums() vocaResponse {
	request, err := http.Get("https://vocadb.net/api/albums?maxResults=5&sort=AdditionDate")
	if err != nil {
		log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		 log.Fatalln(err)
	}
	var jAlbums vocaResponse
	json.Unmarshal(body, &jAlbums)
	jAlbums.TypeOfPull = "Album"
	return jAlbums
}

func pullLatestSongs() vocaResponse {
	request, err := http.Get("https://vocadb.net/api/songs?maxResults=5&sort=AdditionDate")
	if err != nil {
		log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		 log.Fatalln(err)
	}
	var jSongs vocaResponse
	jSongs.TypeOfPull = "Song"
	json.Unmarshal(body, &jSongs)
	return jSongs
}

// CONVERT OBJECT TO RSS

func rsser(latestObject vocaResponse) string {
	now := time.Now()

	typeOfPull := latestObject.TypeOfPull
	var pullApi string
	if typeOfPull == "Album" {
		pullApi = "https://vocadb.net/Al/"
	} else {
		pullApi = "https://vocadb.net/S/"
	}

	feed := &feeds.Feed{
		Title:       "vocadb albums RSS",
		Link:        &feeds.Link{Href: "https://czeczacha.ovh"},
		Description: "vocadb rss feed with latest albums",
		Author:      &feeds.Author{Name: "Jan Maleta"},
		Created:     now,
	}

	for _, vocaItems := range latestObject.Items {
		feed.Items = append(feed.Items,
			&feeds.Item{
				Title:       fmt.Sprintf("%s - %s", vocaItems.Name, vocaItems.ArtistString),
				Link:        &feeds.Link{Href: fmt.Sprint(pullApi, vocaItems.ID)},
				Description: fmt.Sprintf("%s by %s", typeOfPull, vocaItems.Name),
				Author:      &feeds.Author{Name: vocaItems.Name},
				Created:     now,
		  },
		)
	}

	rss, err := feed.ToRss()
	if err != nil {
			log.Fatal(err)
	}
	return rss
}

func httpServer(rssGenSongs string, rssGenAlbums string) {
		// RUN SERVER
		http.HandleFunc("/songs", func(w http.ResponseWriter, r *http.Request){
			fmt.Fprintf(w, rssGenSongs)
		})
		http.HandleFunc("/albums", func(w http.ResponseWriter, r *http.Request){
			fmt.Fprintf(w, rssGenAlbums)
		})
		fmt.Printf("Starting server at port 8080\n")
		if err := http.ListenAndServe(":8080", nil); err != nil {
				log.Fatal(err)
		}
}

func main() {
	latestAlbums := pullLatestAlbums()
	latestSongs := pullLatestSongs()

	rssGenAlbums := rsser(latestAlbums)
	rssGenSongs := rsser(latestSongs)

	go func() {
    for {
			latestAlbums = pullLatestAlbums()
			latestSongs = pullLatestSongs()
			rssGenAlbums = rsser(latestAlbums)
			rssGenSongs = rsser(latestSongs)
			fmt.Println("Refreshed APIs")
			<-time.After(1 * time.Minute)
    }
	}()
	httpServer(rssGenAlbums, rssGenSongs)
}