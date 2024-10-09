.PHONY: build run test cover

build:
	go build -o ./bin/app ./src/cmd/app/main.go

run-test:
ifeq ($(OS),Windows_NT)
	$env:GO_ENV="test"; go run ./src/cmd/app/main.go
else
	GO_ENV=test go run ./src/cmd/app/main.go
endif

run-prod:
ifeq ($(OS),Windows_NT)
    $env:GO_ENV="prod"; go run ./src/cmd/app/main.go
else
	GO_ENV=prod go run ./src/cmd/app/main.go
endif
	
test:
	go test -p 1 -v ./...

cover:
	go test -p 1 -coverprofile="coverage.out" ./...
	go tool cover -html="coverage.out"
