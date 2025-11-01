# Stage 1 — build the Go binary
FROM golang:1.25 AS base-builder

# Install latest Node.js (24.x) and npm for bundling API docs and building v2 frontend
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    ca-certificates \
    gnupg \
  && mkdir -p /etc/apt/keyrings \
  && curl -fsSL https://deb.nodesource.com/gpgkey/nodesource-repo.gpg.key | gpg --dearmor -o /etc/apt/keyrings/nodesource.gpg \
  && echo "deb [signed-by=/etc/apt/keyrings/nodesource.gpg] https://deb.nodesource.com/node_24.x nodistro main" > /etc/apt/sources.list.d/nodesource.list \
  && apt-get update \
  && apt-get install -y --no-install-recommends nodejs \
  && apt-get purge -y gnupg \
  && rm -rf /var/lib/apt/lists/*

FROM base-builder AS builder

RUN go env -w GOCACHE=/go-cache
RUN go env -w GOMODCACHE=/gomod-cache

WORKDIR /src

COPY go.work go.work.sum ./
COPY v1/go.mod v1/go.sum ./v1/
COPY v2/backend/go.mod v2/backend/go.sum ./v2/backend/
COPY v3/go.mod v3/go.sum ./v3/
RUN --mount=type=cache,target=/gomod-cache go mod download

# Install JS deps used for API docs bundling
COPY package.json package-lock.json ./
RUN npm ci

# Add the source tree and build it
ADD . ./
RUN --mount=type=cache,target=/gomod-cache --mount=type=cache,target=/go-cache make v1

# final stage
FROM gcr.io/distroless/base-debian12
COPY --from=builder /src/bin/v1/ /usr/local/bin/