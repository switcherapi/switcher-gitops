build:
	go build -o ./bin/app ./src/cmd/app/main.go

run:
	GOOS=windows $env:GO_ENV="test"; go run ./src/cmd/app/main.go
	GOOS=linux GO_ENV=test go run ./src/cmd/app/main.go
	
test:
	go test -p 1 -coverpkg=./... -v

cover:
	go test -p 1 -coverprofile="coverage.out" ./...
	go tool cover -html="coverage.out"
