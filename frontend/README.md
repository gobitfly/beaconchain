# Beaconch.in Good to know

## Usefull VSC Plugins

- Nuxtr
- EsLint
- Prettier -COde Formatter
- TypeScript Vue Plugin (Volar)
- Vue language Features (Volar)

# Nuxt 3 Minimal Starter

Look at the [Nuxt 3 documentation](https://nuxt.com/docs/getting-started/introduction) to learn more about this framework.

## Setup

Install `npm` and Nuxt.

Clone the `beaconchain` repository from git.

On your console, navigate to folder `beaconchain/frontend`.

Type

```bash
cp .npmrc-example .npmrc
```

In your `.npmrc` file, replace `FA_TOKEN` with an actual key for Font Awesome.

Then type:

```bash
cp .env-example .env
```

In file `.env`, set `NUXT_PRIVATE_SSR_SECRET=<secret>` (replace `<secret>` with the actual secret).

In file `.env`, write the URLs of the API servers and the secret key to access to them.
The variable evoking the development is used to show/hide features and components that are not ready for production.

Set the following mapping in your `/etc/hosts` file:

```
127.0.0.1 local.beaconcha.in
```

Create server certificates for locally running on https, by runing these comands in the console

```bash
openssl genrsa 2048 > server.key
sudo chmod 400 server.key
sudo openssl req -new -x509 -nodes -sha256 -days 365 -key server.key -out server.crt
```

Navigate to folder `beaconchain/frontend` and run

```bash
npm install
```

If you prefer to use _pnpm_, _yarn_ or _bun_ instead of _npm_:

```bash
sudo npm install -g pnpm
sudo npm install -g yarn
sudo npm install -g bun
```

then

```bash
pnpm install
yarn install
bun install
```

## Development Server

Start the development server with one of those commands (they are equivalent, each software having pros and cons) :

```bash
npm run dev
pnpm run dev
yarn dev
bun run dev
```

Now you can browse the front-end at https://local.beaconcha.in:3000/

## Production

Build the application for production:

```bash
# npm
npm run build

# pnpm
pnpm run build

# yarn
yarn build

# bun
bun run build
```

Locally preview production build:

```bash
# npm
npm run preview

# pnpm
pnpm run preview

# yarn
yarn preview

# bun
bun run preview
```

Check out the [deployment documentation](https://nuxt.com/docs/getting-started/deployment) for more information.

## Get mocked api data

If your `user` was added to the `ADMIN` or `DEV` group by the `api team`, you can get
`mocked data` from the `api` for certain `endpoints` by adding `?is_mocked=true` as a 
`query parameter`.

You can `turn on` mocked data `globally` for all `configured enpoints` 
- by setting `NUXT_PUBLIC_IS_API_MOCKED=true`
in your [.env](.env) or
- running `npm run dev:mock:api` (See: [package.json](package.json))

## Descision Record

We documented our decisions in the [decisions](decisions.md) file.
The documentation should be inspired by Architecture Decision Records ([ADR](https://adr.github.io/)).
