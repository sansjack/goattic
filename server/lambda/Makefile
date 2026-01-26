build:
	mkdir -p build
	GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -o build/bootstrap main.go

build-lambda:
	mkdir -p build
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -tags lambda.norpc -trimpath -o build/bootstrap main.go

clean:
	rm -rf build

.PHONY: build build-lambda clean
