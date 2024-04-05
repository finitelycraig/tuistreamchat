GOFILES := $(shell find . -type f -name *.go)

build/tuistreamchat: $(GOFILES) 
	@go build -o build/tuistreamchat cmd/local/main.go

build/tuistreamchatssh: $(GOFILES)
	@go build -o build/tuistreamchatssh cmd/serve/main.go

build-remote: build/tuistreamchatssh-linux-amd64

build/tuistreamchatssh-linux-amd64: $(GOFILES)
	@env GOOS=linux GOARCH=amd64 go build -o build/tuistreamchatssh-linux-amd64 cmd/serve/main.go

.PHONY=run
run: build/tuistreamchat
	@cd build; ./tuistreamchat

.PHONY=serve
serve: build/tuistreamchatssh
	@cd build; ./tuistreamchatssh
