all: build run

all-systray: build-systray run-systray

run:
	@./bin/urlclip

run-systray:
	@./bin/urlclip_systray

build:
	@go build -ldflags="-s -w" -o bin/urlclip cmd/cli/main.go

build-systray:
	@CGO_ENABLED=1 go build -ldflags="-s -w" -o bin/urlclip_systray cmd/systray/main.go
