# Pull in the layer of the base image: alpine:3.10.3
FROM golang:1.11-alpine

WORKDIR /src

COPY *.go go.* ./

RUN CGO_ENABLED=0 go build -o /usr/myappmain

EXPOSE 8010

CMD ["/usr/myappmain"]
