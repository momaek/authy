build:
	GOOS=darwin GOARCH=amd64 go build -o authy-darwin-amd64 main.go

build_docker:
	docker run --rm \
		-v "$$PWD":/authy \
		-v "$$PWD"/.mod:/go/pkg/mod \
		-w /authy \
		-e GOOS=darwin \
		-e GOARCH=amd64 \
		-e CGO_ENABLED=0 \
		-e GOPROXY='https://goproxy.cn,direct' \
		golang:1.15 \
		go build -ldflags="-X 'github.com/momaek/authy/cmd.Version=$$AUTHY_CURRENT_TAG'" -o authy-darwin-amd64 main.go

tar:
	tar zcvf authy.tar.gz authy-darwin-amd64 alfredworkflow/Authy.alfredworkflow

