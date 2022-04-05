FROM golang:1.18.0-alpine3.15

WORKDIR /app

COPY . .

RUN go build -v -o timetrack .

EXPOSE 8000

