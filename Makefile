all:
	go build

install:
	go install

bash:
	go build
	./dexter bash
	sudo mv dexter.sh /etc/bash_completion.d/

test:
	go test -race -cover ./...
