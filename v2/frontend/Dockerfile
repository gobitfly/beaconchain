# this is meant to be built with the root of this repo as build-context
FROM node:18-alpine AS build

RUN mkdir -p /usr/src/nuxt-app
WORKDIR /usr/src/nuxt-app
COPY . .

WORKDIR /usr/src/nuxt-app/frontend
RUN --mount=type=secret,id=NPMRC_FILE,target=/usr/src/nuxt-app/frontend/.npmrc \
    npm ci && npm cache clean --force
RUN npm run build

FROM node:18-alpine
WORKDIR /usr/src/nuxt-app
COPY --from=build /usr/src/nuxt-app/frontend /usr/src/nuxt-app

ENV NUXT_HOST=0.0.0.0
ENV NUXT_PORT=3000

EXPOSE 3000 

ENTRYPOINT ["node", ".output/server/index.mjs"]