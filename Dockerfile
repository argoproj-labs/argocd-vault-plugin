FROM index.docker.io/golang:1.17@sha256:55636cf1983628109e569690596b85077f45aca810a77904e8afad48b49aa500

ADD go.mod go.mod
ADD go.sum go.sum

ENV GOPATH=""
RUN go mod download

VOLUME work
WORKDIR work
