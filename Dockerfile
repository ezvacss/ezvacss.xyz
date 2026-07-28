# Build stage
FROM golang:1.26.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Run stage
FROM alpine:3.19

RUN adduser -D -u 1000 appuser
WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/public ./public
COPY --from=builder /app/db/schema.sql ./db/schema.sql

RUN chown -R appuser:appuser /app
USER appuser

ENV PORT=8080
EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --retries=3 \
  CMD wget -qO- http://localhost:8080/healthz || exit 1

CMD ["./main"]