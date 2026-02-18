build:
	mkdir -p build
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -trimpath -o build/bootstrap main.go

clean:
	rm -rf build

.PHONY: build clean
