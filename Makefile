ARM = arm
MY_ARCH = $(shell go env GOARCH)

.PHONY: all
all: analyze build test 

.PHONY: build
build:
	go fmt
	go build

.PHONY: analyze
analyze:
	golint
	go vet
	go fmt
	gosec ./...

.PHONY: test
test:
	export GODEBUG="x509ignoreCN=0"

# cannot use "-race" flag on ARM systems
ifeq ($(MY_ARCH), $(ARM))
	sudo GODEBUG="x509ignoreCN=0" go test -v  -coverprofile=coverage.out -p 1
else 
	sudo GODEBUG="x509ignoreCN=0" go test -v -race -coverprofile=coverage.out -p 1
endif
	go tool cover -html coverage.out -o coverage.html

.PHONY: clean
clean: 
ifneq ("$(wildcard coverage.html)", "")
	rm -f coverage.html
endif
ifneq ("$(wildcard coverage.out)", "")
	rm -f coverage.out
endif
	go clean



