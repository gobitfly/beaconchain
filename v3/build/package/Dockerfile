# Stage 1 — build the Go binary
FROM golang:1.23 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server cmd/main.go

# Stage 2 — minimal runtime image
FROM gcr.io/distroless/static-debian11

WORKDIR /

COPY --from=builder /app/server .
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/api/gen ./api/gen
COPY --from=builder /app/web ./web

# Cloud Run expects the service to listen on PORT (8080 by default)
ENV PORT=8080

# Start the app
ENTRYPOINT ["/server"]