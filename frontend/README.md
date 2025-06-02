# Beaconcha.in Good to know

## Usefull VSC Plugins

- Nuxtr
- EsLint
- Prettier -COde Formatter
- TypeScript Vue Plugin (Volar)
- Vue language Features (Volar)

# Nuxt 3 Minimal Starter

Look at the [Nuxt 3 documentation](https://nuxt.com/docs/getting-started/introduction) to learn more about this framework.

## Setup

Clone the `beaconchain` repository from git.

In your terminal, navigate to folder `beaconchain/frontend`.

### Install packages.

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

### Create `.env` file
Copy the existing `.env-example` file to a new `.env`:

```bash
cp .env-example .env
```

Inside of `.env`, add necessary URLs and secrets.

### Create `SSL certificate` to enable local development over `https`

We recommend using `mkcert` for creating self-signed certificates, as it simplifies the process and avoids browser warnings. If you don't have `mkcert` installed, you can follow the [installation instructions](https://github.com/FiloSottile/mkcert#installation).

Once installed, run:

```bash
# create a local CA (Certificate Authority)
mkcert -install
# create a certificate for your local development domain
mkcert local.beaconcha.in
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

## Decision Record

We documented our decisions in the [decisions](decisions.md) file.
The documentation should be inspired by Architecture Decision Records ([ADR](https://adr.github.io/)).
