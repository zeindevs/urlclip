all: build run

run:
	@./bin/urlclip

build:
	@go build -o bin/urlclip main.go
