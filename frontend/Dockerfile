# this is meant to be built with the root of this repo as build-context
FROM node:22 AS build

RUN mkdir -p /usr/src/nuxt-app
WORKDIR /usr/src/nuxt-app
COPY . .

WORKDIR /usr/src/nuxt-app
RUN --mount=type=secret,id=NPMRC_FILE,target=/usr/src/nuxt-app/.npmrc \
    npm ci && npm cache clean --force
RUN npm run build

FROM gcr.io/distroless/nodejs22-debian12
WORKDIR /usr/src/nuxt-app
COPY --from=build /usr/src/nuxt-app /usr/src/nuxt-app

ARG BEACONCHAIN_VERSION
ENV NUXT_PUBLIC_GIT_VERSION=$BEACONCHAIN_VERSION
ENV NUXT_HOST=0.0.0.0
ENV NUXT_PORT=3000

EXPOSE 3000 

CMD [".output/server/index.mjs"]