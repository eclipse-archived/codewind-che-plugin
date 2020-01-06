PKGS := $(shell go list  ./... | grep -v $(PROJECT)/vendor | grep -v $(PROJECT)/tests )
BUILD_FLAGS := -ldflags="-s -w"

default: bin

.PHONY: bin
bin:
	dep ensure -v
	CGO_ENABLED=0 go build ${BUILD_FLAGS} -o deploy-pfe

.PHONY: test
test:
	go test -race $(PKGS)