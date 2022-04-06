FROM golang:1.18.0-alpine3.15

RUN apk update && apk add --no-cache git && apk add --no-cach bash && apk add build-base

WORKDIR /app

COPY . .

RUN go build -v -o timetrack .

EXPOSE 8000

