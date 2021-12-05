FROM golang:alpine
LABEL maintainer="Aliucord"

WORKDIR /build

RUN apk add git

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY main.go .
COPY bot bot
COPY common common
COPY updateTracker updateTracker
COPY database database

RUN go build -o aliucord-backend

WORKDIR /app
RUN cp /build/aliucord-backend .
COPY config.json .

RUN adduser --disabled-password --gecos "" juby
USER juby
CMD [ "/app/aliucord-backend" ]
