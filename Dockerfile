FROM golang:1.16.5-alpine3.13

WORKDIR /app

COPY . .

RUN go build -v -o timetrack .

EXPOSE 8000

