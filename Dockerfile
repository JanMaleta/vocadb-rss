FROM alpine:3.15.4
RUN apk --no-cache add ca-certificates

COPY vocadbRSS /

EXPOSE 8080
CMD ["/vocadbRSS"]