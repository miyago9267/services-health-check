# syntax=docker/dockerfile:1
FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags "-s -w" -o /out/healthd ./cmd/healthd

FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache tzdata \
    && cp /usr/share/zoneinfo/Asia/Taipei /etc/localtime \
    && echo "Asia/Taipei" > /etc/timezone \
    && addgroup -S healthd && adduser -S -G healthd healthd \
    && mkdir -p /etc/healthd \
    && chown -R healthd:healthd /app /etc/healthd
ENV TZ=Asia/Taipei
COPY --from=builder /out/healthd /app/healthd
COPY configs /etc/healthd
USER healthd

EXPOSE 8080
ENTRYPOINT ["/app/healthd"]
CMD ["-config", "/etc/healthd/example.yaml"]
