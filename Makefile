build:
	go build -o bin/dslbot .

deps:
	go get -v github.com/gorilla/mux
	go get -v github.com/urfave/negroni
crypto:
	openssl genrsa -out server.key 2048
	openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650

deploy: deps build
all: deps crypto build
