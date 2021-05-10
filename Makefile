BINARY=argocd-vault-plugin
VERSION=0.7.1
OS_ARCH=darwin_amd64

default: build

quality:
	go vet github.com/IBM/argocd-vault-plugin/...
	go test -v -coverprofile cover.out ./...

build:
	go build -o ${BINARY} .

release:
	GOOS=darwin GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_darwin_amd64
	GOOS=linux GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_linux_amd64

install: build

e2e: install
	./argocd-vault-plugin
