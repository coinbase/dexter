all:
	go build

install:
	go install

bash:
	go build
	./dexter bash
	sudo mv dexter.sh /etc/bash_completion.d/

deps:
	go get github.com/golang/dep/cmd/dep
	dep ensure

test:
	go test -race -cover `go list ./... | grep -v /vendor/`
