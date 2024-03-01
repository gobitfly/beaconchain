FROM node:18-alpine AS build

RUN mkdir -p /usr/src/nuxt-app
WORKDIR /usr/src/nuxt-app
COPY . .

RUN echo "//npm.fontawesome.com/:_authToken=${FA_PRO_KEY}" > .npmrc && npm ci && npm cache clean --force && rm -f .npmrc
RUN npm run build

FROM node:18-alpine
WORKDIR /usr/src/nuxt-app
COPY --from=build /usr/src/nuxt-app /usr/src/nuxt-app

ENV NUXT_HOST=0.0.0.0
ENV NUXT_PORT=3000

EXPOSE 3000 

ENTRYPOINT ["node", ".output/server/index.mjs"]