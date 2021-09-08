FROM golang:rc-buster as builder

WORKDIR /go/src/app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -v -o wedeo-app ./app/main.go


FROM debian:stable-slim

RUN apt-get update && \
    apt-get -y upgrade && \
    apt-get install -y ffmpeg && \
    apt-get install -y ca-certificates apache2-utils && \
    rm -rf /var/cache/apt/*

WORKDIR /go/src/app

COPY --from=builder /go/src/app/ /go/src/app/

CMD ["./wedeo-app"]
