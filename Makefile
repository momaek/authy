build:
	GOOS=darwin GOARCH=amd64 go build -o authy-export-darwin-amd64 main.go
	GOOS=darwin GOARCH=386 go build -o authy-export-darwin-386 main.go