FROM golang:1.14

WORKDIR /go/src/dslmailer

RUN go get -v github.com/gorilla/mux
RUN go get -v github.com/urfave/negroni

RUN mkdir /cfg

#COPY cfg.json /cfg/cfg.json
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

#RUN make deploy
#RUN go build -o dslmailer .

CMD ["/go/bin/dslmailer", "start", "/cfg/cfg.json"]