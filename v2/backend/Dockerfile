# this is meant to be built with the root of this repo as build-context
FROM golang:1.23.0 AS build-env
COPY backend/go.mod backend/go.sum /src/backend/
WORKDIR /src/backend
RUN go mod download
ADD . /src
ARG target=all
RUN make -B $target

# final stage
FROM gcr.io/distroless/base-debian12
COPY --from=build-env /src/backend/bin /usr/local/bin/
CMD ["bc","exporter"]