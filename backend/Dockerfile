# this is meant to be built with the root of this repo as build-context
FROM golang:1.22.0 AS build-env
COPY backend/go.mod backend/go.sum /src/backend/
WORKDIR /src/backend
RUN go mod download
ADD . /src
ARG target=all
RUN make -B $target

# final stage
FROM ubuntu:22.04
RUN apt-get update && apt-get -y upgrade && apt-get install -y --no-install-recommends \
  libssl-dev \
  ca-certificates \
  && apt-get clean \
  && rm -rf /var/lib/apt/lists/*
WORKDIR /usr/local/bin
COPY --from=build-env /src/backend/bin /usr/local/bin/
CMD ["./exporter"]