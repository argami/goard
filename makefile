REVISION = $(shell git rev-parse HEAD)
HOSTNAME = $(shell hostname -s)

darwin:
	go build -ldflags "-X main.revision=$(REVISION)-$(HOSTNAME)" goard.go

linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.revision=$(REVISION)-$(HOSTNAME)" goard.go
