# Stage 1 — build the Go binary
FROM golang:1.25 AS base-builder

# Install all optional dependencies for building here
# Install buf (used by v3)
ENV BUF_BIN=/usr/local/bin
ENV BUF_VERSION=1.55.1
RUN curl -sSL \
    "https://github.com/bufbuild/buf/releases/download/v${BUF_VERSION}/buf-$(uname -s)-$(uname -m)" \
    -o "${BUF_BIN}/buf" && \
    chmod +x "${BUF_BIN}/buf"

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
RUN --mount=type=cache,target=/gomod-cache --mount=type=cache,target=/go-cache make v3

# final stage
FROM gcr.io/distroless/base-debian12
COPY --from=builder /src/bin/v3/ /usr/local/bin/