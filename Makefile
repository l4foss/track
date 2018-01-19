TARGET = track
VERSION = $(shell git rev-parse --short HEAD)
LDFLAGS = -X main.REVISION="$(VERSION)"

all:
	go build -ldflags "$(LDFLAGS)" -o $(TARGET)
clean:
	rm -f $(TARGET)
