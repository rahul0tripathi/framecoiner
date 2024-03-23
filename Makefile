build-common:
	@ go version
	@ go clean
	@ go mod tidy && go mod download
	@ go mod verify

build: build-common
	@ CGO_ENABLED=0 go build -ldflags='-w -s -extldflags "-static"' -a -o "_bin/framecoiner" cmd/*.go