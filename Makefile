BINARY=argocd-vault-plugin

default: build

quality:
	go vet github.com/IBM/argocd-vault-plugin/...
	go test -v -coverprofile cover.out ./...

build:
	go build -o ${BINARY} .

release:
	goreleaser release --skip-publish --rm-dist

install: build

e2e: install
	./argocd-vault-plugin
