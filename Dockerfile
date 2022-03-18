# golang:1.17-alpine
FROM golang@sha256:ecd7f25b110a6b9b3b0b9f29faf1cd61e216569b250cf5ec4c4b044a8ee2d366

RUN apk add make git

ADD go.mod go.mod
ADD go.sum go.sum

ENV GOPATH=""
ENV CGO_ENABLED=0
RUN go mod download

VOLUME work
WORKDIR work