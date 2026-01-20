# syntax=docker/dockerfile:1
FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/healthd ./cmd/healthd

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /out/healthd /app/healthd
COPY configs /app/configs

EXPOSE 8080
ENTRYPOINT ["/app/healthd"]
