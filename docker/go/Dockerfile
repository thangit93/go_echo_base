# Build stage
FROM golang:1.24.3-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download \
  && go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api/main.go

# Run stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]