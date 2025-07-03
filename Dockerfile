FROM --platform=linux/arm64 golang:1.23.2 AS builder

WORKDIR /app

# cache dependencies
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a -ldflags='-s -w' -o server ./main.go

FROM --platform=linux/arm64 debian:bullseye-slim

WORKDIR /app

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY server.cfg.yaml .
COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]