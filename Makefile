BINARY=argocd-vault-plugin

default: build

quality:
	go vet github.com/IBM/argocd-vault-plugin
	go test -v -coverprofile cover.out ./...

build:
	go build -o ${BINARY} .

install: build

e2e: install
	./argocd-vault-plugin
