FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o myapp ./cmd/api/main.go

FROM alpine:3.19

WORKDIR /

COPY --from=builder /app/myapp /myapp

ENTRYPOINT ["/myapp"]