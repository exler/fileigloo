# syntax=docker/dockerfile:1
ARG GO_VERSION=1.20

FROM golang:${GO_VERSION}-alpine as build_go

RUN apk add git 

WORKDIR /app
COPY . /app

ENV GO111MODULE=on
ENV CGO_ENABLED=0

RUN go build -tags urfave_cli_no_docs -ldflags "-X github.com/exler/fileigloo/cmd.Version=$(git describe --tags)" -o /fileigloo

FROM alpine:3.18

WORKDIR /app
COPY --from=build_go /fileigloo /app/fileigloo

ENTRYPOINT ["/app/fileigloo", "runserver"]

EXPOSE 8000

