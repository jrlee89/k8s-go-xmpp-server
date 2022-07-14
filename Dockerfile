FROM golang:alpine
EXPOSE 5222
RUN mkdir /app
COPY . /app
WORKDIR /app
RUN go build main.go client.go server.go ns.go rng.go stanza.go tee.go
CMD ["/app/main"]
