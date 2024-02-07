build:
	GOOS=darwin GOARCH=arm64 go build -o output/bitrise-plugins-swift-Darwin-arm64
	GOOS=darwin GOARCH=amd64 go build -o output/bitrise-plugins-swift-Darwin-x86_64
	GOOS=linux GOARCH=amd64 go build -o output/bitrise-plugins-swift-Linux-x86_64
