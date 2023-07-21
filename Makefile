Version=$$(git describe --tags)
build_x86:
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X github.com/momaek/authy/cmd.Version=${Version}" -o authy-darwin-amd64 main.go

build_m1:
	GOOS=darwin GOARCH=arm64 go build -ldflags "-X github.com/momaek/authy/cmd.Version=${Version}" -o authy-darwin-arm64 main.go

build: build_x86 build_m1

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
	tar zcvf authy.tar.gz authy-darwin-amd64 authy-darwin-arm64 alfredworkflow/Authy.alfredworkflow

