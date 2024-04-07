GOFILES := $(shell find . -type f -name *.go)

tidy:
	@go mod tidy

fmt: $(GOFILES)
	@echo 'Formatting go files'
	@go fmt cmd/local/main.go
	@go fmt cmd/serve/main.go
	@go fmt data/*.go
	@go fmt internal/*.go

build/tuistreamchat: $(GOFILES) fmt tidy
	@echo 'Building local binary file build/tuistreamchat'
	@go build -o build/tuistreamchat cmd/local/main.go

build/tuistreamchatssh: $(GOFILES) fmt tidy
	@echo 'Building ssh binary file build/tuistreamchatssh'
	@go build -o build/tuistreamchatssh cmd/serve/main.go

build-remote: build/tuistreamchatssh-linux-amd64

build/tuistreamchatssh-linux-amd64: $(GOFILES) fmt tidy
	@echo 'Building ssh binary file build/tuistreamchatssh-linux-amd64'
	@env GOOS=linux GOARCH=amd64 go build -o build/tuistreamchatssh-linux-amd64 cmd/serve/main.go

.PHONY=run
run: build/tuistreamchat
	@cd build; ./tuistreamchat

.PHONY=serve
serve: build/tuistreamchatssh
	@echo 'Serving tuistreamchat over ssh'
	@cd build; ./tuistreamchatssh

