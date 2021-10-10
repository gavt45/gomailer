FROM golang:1.16

ENV SMTP_ADDR=smtp.gmail.com
ENV SMTP_PASSWORD=secret
ENV EMAIL=test@gmail.com
ENV PORT=80
ENV SECRET=secret
ENV TEMPLATE_PATH=/templates
ENV GOBIN=/usr/local/bin/


WORKDIR /go/src/dslmailer

COPY go.mod .

RUN go mod download

RUN mkdir /templates

#COPY cfg.json /cfg/cfg.json
COPY *.go ./
COPY Makefile ./

RUN go get -d -v ./...
RUN go install -v ./...

#RUN make deploy
#RUN go build -o dslmailer .

CMD ["/usr/local/bin/gomailer", "start"]