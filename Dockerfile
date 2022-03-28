# golang:1.17-alpine
FROM golang@sha256:f4ece20984a30d1065b04653bf6781f51ab63421b4b8f011565de0401cfe58d7

RUN apk add make git

ADD go.mod go.mod
ADD go.sum go.sum

ENV GOPATH=""
ENV CGO_ENABLED=0
RUN go mod download

VOLUME work
WORKDIR work
