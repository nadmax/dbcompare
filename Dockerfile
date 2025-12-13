FROM golang:1.25.3-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY ./ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o dbcompare ./cmd/dbcompare/

FROM alpine:3.22 AS final
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /build/dbcompare ./
RUN mkdir -p /app/configs /app/results && \
    adduser -D -u 1000 user && \
    chown -R user:user /app
USER user
ENTRYPOINT ["/app/dbcompare"]
CMD ["-config", "/app/configs/config.yaml"]
