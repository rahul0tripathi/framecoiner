build-common:
	@ go version
	@ go clean
	@ go mod tidy && go mod download
	@ go mod verify

build: build-common
	@ go build -o "_bin/framecoiner" cmd/main.go