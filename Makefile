BINARY=argocd-vault-plugin

.PHONY: default
default: build

.PHONY: quality
quality: vet lint test

.PHONY: vet
vet:
	go vet github.com/argoproj-labs/argocd-vault-plugin

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: test
test:
	go test -v -coverprofile cover.out ./...

.PHONY: build
build:
	go build -o ${BINARY} .

.PHONY: install
install: build

.PHONY: e2e
e2e: install
	./argocd-vault-plugin
