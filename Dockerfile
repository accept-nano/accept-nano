FROM golang:1.14

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

# entry shell
COPY docker/entry.sh /entry.sh

# go for it!
CMD ["/bin/bash", "/entry.sh"]

EXPOSE 8080