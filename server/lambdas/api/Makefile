build:
	mkdir -p build
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -trimpath -o build/bootstrap main.go

dev:
	AWS_ENDPOINT_URL=http://localhost:4566 AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test AWS_REGION=eu-west-2 go run main.go

clean:
	rm -rf build localstack-data

.PHONY: build dev clean
