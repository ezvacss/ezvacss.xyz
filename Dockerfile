# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy code files and web directory
COPY . .

# Build statically linked Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Run stage
FROM alpine:3.19

WORKDIR /app

# Copy executable and supporting directories
COPY --from=builder /app/main .
COPY --from=builder /app/public ./public
COPY --from=builder /app/db/schema.sql ./db/schema.sql

ENV PORT=8080
EXPOSE 8080

CMD ["./main"]
