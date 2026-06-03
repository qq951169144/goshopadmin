FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /order-service ./cmd/server

FROM alpine:latest
WORKDIR /app
COPY --from=builder /order-service .
EXPOSE 8090
CMD ["./order-service"]
