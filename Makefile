.PHONY: build clean deploy

build:
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o bootstrap main.go

clean:
	rm -f bootstrap

deploy: clean build
	npx serverless deploy