FROM node:18-alpine AS build

RUN mkdir -p /usr/src/nuxt-app
WORKDIR /usr/src/nuxt-app
COPY . .

RUN --mount=type=secret,id=NPMRC_FILE,target=/usr/src/nuxt-app/.npmrc \
    npm ci && npm cache clean --force
RUN npm run build

FROM node:18-alpine
WORKDIR /usr/src/nuxt-app
COPY --from=build /usr/src/nuxt-app /usr/src/nuxt-app

ENV NUXT_HOST=0.0.0.0
ENV NUXT_PORT=3000

EXPOSE 3000 

ENTRYPOINT ["node", ".output/server/index.mjs"]