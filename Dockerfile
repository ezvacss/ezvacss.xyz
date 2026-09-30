#build
FROM golang:1.27.1-alpine as build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /main ./cmd/server

#test
FROM build-stage as run-test-stage
RUN go test -v ./...

#build release
FROM alpine:3.23.5 AS build-release-stage

WORKDIR /

RUN adduser -D -u 1111 nonroot

USER nonroot

WORKDIR /app
COPY --from=build-stage --chown=nonroot:nonroot /app/public ./public
COPY --from=build-stage --chown=nonroot:nonroot /main ./main

EXPOSE 8080

ENTRYPOINT ["./main"]