BINARY=argocd-vault-plugin

default: build

quality:
	go vet github.com/argoproj-labs/argocd-vault-plugin
	go test -race -v -coverprofile cover.out ./...

build:
	go build -o ${BINARY} .

install: build

e2e: install
	./argocd-vault-plugin
