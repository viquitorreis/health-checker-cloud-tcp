.PHONY: build-arm
build-arm:
	@GOOS=linux GOARCH=arm go build -o bin/healthchecker .

.PHONY: build
build:
	@go build -o bin/healthchecker .
