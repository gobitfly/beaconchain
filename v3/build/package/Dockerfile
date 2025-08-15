# Stage 1 — build the Go binary
FROM golang:1.23 AS builder
WORKDIR /app

# Install buf
ENV BIN=/usr/local/bin
ENV VERSION=1.55.1
RUN curl -sSL \
    "https://github.com/bufbuild/buf/releases/download/v${VERSION}/buf-$(uname -s)-$(uname -m)" \
    -o "${BIN}/buf" && \
    chmod +x "${BIN}/buf"

COPY go.mod go.sum ./
RUN go mod download

# TODO remove after config handling is implemented
COPY configs/service/default.yaml.example configs/service/default.yaml
COPY configs/service/personal_cloudrun.yaml.example configs/service/personal_cloudrun.yaml

COPY . .
RUN make generate-proto
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server cmd/main.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go test -c -o tests ./test

# Stage 2 — minimal runtime image
FROM gcr.io/distroless/static-debian11

WORKDIR /

COPY --from=builder /app/server .
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/api/gen ./api/gen
COPY --from=builder /app/web ./web

# Tests
COPY --from=builder /app/tests .
COPY --from=builder /app/test/.env.example /test/.env.example

# Cloud Run expects the service to listen on PORT (8080 by default)
ENV PORT=8080

# Start the app
ENTRYPOINT ["/server"]