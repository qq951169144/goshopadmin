FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /product-service ./cmd/server

FROM alpine:latest
WORKDIR /app
COPY --from=builder /product-service .
EXPOSE 8091
CMD ["./product-service"]
