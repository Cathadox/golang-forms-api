FROM golang:1.21-alpine

RUN apk add --no-cache git

RUN go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

ENTRYPOINT ["/go/bin/oapi-codegen"]