TARGET = track

all:
	go build -o $(TARGET)
clean:
	rm -f $(TARGET)
