#build
FROM golang:1.26-alpine as build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /main

#test
FROM build-stage as run-test-stage
RUN go test -v ./...

#build release
FROM alpine:3.23.5 AS build-release-stage

WORKDIR /

RUN adduser -D -u 1111 nonroot

COPY --from=build-stage --chown=nonroot:nonroot /main /main

USER nonroot

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --retries=3 \
    CMD wget -qO- http://localhost:8080/healthz || exit 1

ENTRYPOINT ["/main"]