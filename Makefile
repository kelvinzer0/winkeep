VERSION ?= $(shell date +%Y.%m.%d)
LDFLAGS = -s -w -X main.version=$(VERSION)

.PHONY: build build-all clean test fmt vet

build:
	go build -ldflags="$(LDFLAGS)" -o bin/winkeep.exe ./cmd/winkeep/

build-all:
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/winkeep_windows_amd64.exe ./cmd/winkeep/
	GOOS=windows GOARCH=386 go build -ldflags="$(LDFLAGS)" -o dist/winkeep_windows_386.exe ./cmd/winkeep/
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/winkeep_linux_amd64 ./cmd/winkeep/
	GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o dist/winkeep_linux_arm64 ./cmd/winkeep/
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/winkeep_darwin_amd64 ./cmd/winkeep/
	GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o dist/winkeep_darwin_arm64 ./cmd/winkeep/

clean:
	rm -rf bin/ dist/

test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

run: build
	./bin/winkeep.exe version

release: build-all
	@echo "Build complete: dist/"
	@ls -la dist/
