FROM golang:1.14.4-alpine3.12 AS builder

WORKDIR /go/src/accept-nano

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go install

###############################################################################

FROM alpine:3.12.0

COPY --from=builder /go/bin/accept-nano /usr/bin/accept-nano
RUN chmod +x /usr/bin/accept-nano

COPY docker ./docker

CMD ["/bin/sh", "/docker/entry.sh"]

EXPOSE 8080
