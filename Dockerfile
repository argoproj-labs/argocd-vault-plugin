FROM golang@sha256:8474650232fca6807a8567151ee0a6bd2a54ea28cfc93f7824b42267ef4af693

RUN apk add make git

ADD go.mod go.mod
ADD go.sum go.sum

ENV GOPATH=""
ENV CGO_ENABLED=0
RUN go mod download

VOLUME work
WORKDIR work