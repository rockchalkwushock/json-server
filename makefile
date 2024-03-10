run: build
	@./bin/app

build:
	@go build -v -o ./bin/app main.go
