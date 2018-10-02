all: embed

install: embed
	go install

bash:
	go build
	./dexter bash
	sudo mv dexter.sh /etc/bash_completion.d/

deps:
	go get github.com/UnnoTed/fileb0x
	go get github.com/golang/dep/cmd/dep
	dep ensure
	mkdir -p investigators

embed:
	fileb0x embedded/filebox.yml
	go build

test:
	go test -race -cover `go list ./... | grep -v /vendor/`
