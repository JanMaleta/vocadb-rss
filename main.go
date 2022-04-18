package main

import (
  "fmt"
  "time"
  "io/ioutil"
  "log"
  "net/http"
  "encoding/json"
  "github.com/gorilla/feeds"
)

type vocaResponse struct {
  Items[] vocaItems `json:"items"`
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
  PublishDate string `json:"publishDate",omitempty` //Returned only in songs
}
type albumsReleaseDate struct {
  Formatted string `json:"formatted",omitempty` //Returned only in albums
}

func pullLatestData(apiUrl string, typeOfPull string) vocaResponse {
  request, err := http.Get(apiUrl)
  if err != nil {
    log.Fatalln(err)
  }
  body, err := ioutil.ReadAll(request.Body)
  if err != nil {
    log.Fatalln(err)
  }
  var jAlbums vocaResponse
  json.Unmarshal(body, & jAlbums)
  jAlbums.TypeOfPull = typeOfPull
  return jAlbums
}

func rsser(latestObject *vocaResponse) string {
  now := time.Now()

	typeOfPull := latestObject.TypeOfPull
  var pullApi string
  if typeOfPull == "Album" {
    pullApi = "https://vocadb.net/Al/"
  } else {
    pullApi = "https://vocadb.net/S/"
  }

  feed := &feeds.Feed {
    Title: "vocadb albums RSS",
    Link: &feeds.Link { Href: "https://czeczacha.ovh" },
    Description: "vocadb rss feed with latest albums",
    Author: &feeds.Author { Name: "Jan Maleta" },
    Created: now,
  }

  for _, vocaItems:= range latestObject.Items {
    timeParsed, err := time.Parse("2006-01-02T15:04:05.999", vocaItems.CreateDate)
    if err != nil {
      fmt.Println(err)
    }
    feed.Items = append(feed.Items, & feeds.Item {
      Title: fmt.Sprintf("%s - %s", vocaItems.ArtistString, vocaItems.Name),
      Link: &feeds.Link { Href: fmt.Sprint(pullApi, vocaItems.ID) },
      Description: fmt.Sprintf("%s by %s", typeOfPull, vocaItems.ArtistString),
      Author: &feeds.Author { Name: vocaItems.ArtistString },
      Created: timeParsed,
    }, )
  }

  rss, err := feed.ToRss()
    if err != nil {
      log.Fatal(err)
    }
  return rss
}

func httpServer(rssGenSongs *string, rssGenAlbums *string) {
  http.HandleFunc("/songs", func(w http.ResponseWriter, r * http.Request) {
    fmt.Fprintf(w, *rssGenSongs)
  })
  http.HandleFunc("/albums", func(w http.ResponseWriter, r * http.Request) {
    fmt.Fprintf(w, *rssGenAlbums)
  })
  fmt.Printf("Starting server at port 8080\n")
  if err := http.ListenAndServe(":8080", nil);
  err != nil {
    log.Fatal(err)
  }
}

func main() {
  songsUrl := "https://vocadb.net/api/songs?maxResults=50&sort=AdditionDate"
  albumsUrl := "https://vocadb.net/api/albums?maxResults=50&sort=AdditionDate"

  latestSongs := pullLatestData(songsUrl, "Song")
  pointerLatestSongs := &latestSongs
  latestAlbums := pullLatestData(albumsUrl, "Album")
  pointerLatestAlbums := &latestAlbums

  rssGenSongs := rsser(pointerLatestSongs)
  pointerRssGenSongs := &rssGenSongs
	rssGenAlbums := rsser(pointerLatestAlbums)
  pointerRssGenAlbums := &rssGenAlbums

  go func() {
    for {
      latestSongs = pullLatestData(songsUrl, "Song")
      latestAlbums = pullLatestData(albumsUrl, "Album")
      rssGenSongs = rsser(pointerLatestSongs)
      rssGenAlbums = rsser(pointerLatestAlbums)
      <-time.After(5 * time.Minute)
    }
  }()
  httpServer(pointerRssGenSongs, pointerRssGenAlbums)
}