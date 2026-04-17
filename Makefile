.PHONY: build tidy run clean

BINARY_NAME=golink

build: tidy
	CGO_ENABLED=1 go build -o $(BINARY_NAME) ./cmd/golink

tidy:
	go mod tidy

run: build
	./$(BINARY_NAME) --config config.yml

clean:
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME).exe
