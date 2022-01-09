# syntax=docker/dockerfile:1

FROM golang:1.17-alpine

WORKDIR /app

COPY ./.env ./
COPY ./stats ./

EXPOSE 80

CMD [ "./stats" ]