FROM index.docker.io/golang:1.22.7

ADD go.mod go.mod
ADD go.sum go.sum

ENV GOPATH=""
RUN go mod download

VOLUME work
WORKDIR work
