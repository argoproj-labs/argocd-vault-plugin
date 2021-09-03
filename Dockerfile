FROM golang@sha256:a92abf53d0ac8dd292b432525c15ce627554546ebc4164d1d4149256998122ce

RUN apk add make

ADD go.mod go.mod
ADD go.sum go.sum

ENV GOPATH=""
ENV CGO_ENABLED=0
RUN go mod download

VOLUME work
WORKDIR work