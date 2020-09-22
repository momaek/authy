build:
	GOOS=darwin GOARCH=amd64 go build -o authy-export-darwin-amd64 main.go

build_docker:
	docker run --rm \
		-v "$$PWD":/authy \
		-v "$$PWD"/.mod:/go/pkg/mod \
		-w /authy \
		-e GOOS=darwin \
		-e GOARCH=amd64 \
		-e CGO_ENABLED=0 \
		-e GOPROXY='https://goproxy.cn,direct' \
		docker.elastic.co/beats-dev/golang-crossbuild:1.14.7-darwin \
		--build-cmd "go build -o authy-darwin-amd64 main.go" \
		-p 'darwin/amd64'
