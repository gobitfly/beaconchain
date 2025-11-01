# Stage 1 — build the Go binary
FROM golang:1.25 AS base-builder

# Install all optional dependencies for building here

FROM base-builder AS builder

RUN go env -w GOCACHE=/go-cache
RUN go env -w GOMODCACHE=/gomod-cache

WORKDIR /src

COPY go.work go.work.sum ./
COPY v1/go.mod v1/go.sum ./v1/
COPY v2/backend/go.mod v2/backend/go.sum ./v2/backend/
COPY v3/go.mod v3/go.sum ./v3/
RUN --mount=type=cache,target=/gomod-cache go mod download

# Add the source tree and build it
ADD . ./
RUN --mount=type=cache,target=/gomod-cache --mount=type=cache,target=/go-cache make v2

# final stage
FROM gcr.io/distroless/base-debian12
COPY --from=builder /src/bin/v2/ /usr/local/bin/