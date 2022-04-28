build:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0  go build -ldflags='-s' -o armhandler .
	GOOS=linux GOARCH=amd64 go build -o ./handler
	zip -j ./testlambda.zip ./handler
