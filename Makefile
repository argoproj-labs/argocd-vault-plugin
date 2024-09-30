BINARY=argocd-vault-plugin
LD_FLAGS=-ldflags '-linkmode external -extldflags "-static -Wl,-unresolved-symbols=ignore-all"'

default: build

quality:
	go vet github.com/argoproj-labs/argocd-vault-plugin
	go test ${LD_FLAGS} -race -v -coverprofile cover.out ./...

build:
	C=musl-gcc CGO_ENABLED=1 go build ${LD_FLAGS} -buildvcs=false -o ${BINARY} .

test:
	C=musl-gcc CGO_ENABLED=1 go test ${LD_FLAGS} -buildvcs=false ./...
	
install: build

e2e: install
	./argocd-vault-plugin
