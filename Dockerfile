FROM golang:1.14

WORKDIR /go/src/app

COPY go.mod go.sum ./
RUN go mod download -x

COPY . .
RUN go install -v

# entry shell
COPY docker/entry.sh /entry.sh

# go for it!
CMD ["/bin/bash", "/entry.sh"]

EXPOSE 8080
