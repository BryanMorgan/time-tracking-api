FROM golang:1.15.3-alpine3.12

WORKDIR /app

COPY . .

RUN go build -v -o timetrack .

EXPOSE 8000

