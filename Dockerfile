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

RUN mkdir -p /home/juby/Aliucord-backend
WORKDIR /home/juby/Aliucord-backend
RUN cp /build/aliucord-backend .
COPY config.json .

RUN adduser --disabled-password --gecos "" juby
USER juby

CMD [ "/home/juby/Aliucord-backend/aliucord-backend" ]
