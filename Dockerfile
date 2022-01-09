# syntax=docker/dockerfile:1

FROM golang:1.17-alpine

WORKDIR /app

COPY .env ./
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o /users

EXPOSE 80

CMD [ "/users" ]