FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /customer-service ./cmd/server

FROM alpine:latest
WORKDIR /app
COPY --from=builder /customer-service .
EXPOSE 8092
CMD ["./customer-service"]
