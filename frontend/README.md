# Beaconch.in Good to know

## Usefull VSC Plugins
- Nuxtr
- EsLint
- Prettier -COde Formatter
- TypeScript Vue Plugin (Volar)
- Vue language Features (Volar)

# Nuxt 3 Minimal Starter

Look at the [Nuxt 3 documentation](https://nuxt.com/docs/getting-started/introduction) to learn more.

## Setup

Set the following mapping in your `/etc/hosts` file
`127.0.0.1 local.beaconcha.in`

Create server certificates for locally running on https, by runing this comands in the console
```bash
openssl genrsa 2048 > server.key
chmod 400 server.key
openssl req -new -x509 -nodes -sha256 -days 365 -key server.key -out server.crt
```
Set the following env variable (needed to load local mock data): 
`export NODE_TLS_REJECT_UNAUTHORIZED=0`

Make sure to install the dependencies:

copy .npmrc-example to .npmrc and replace YOURKEY with your fontawesome API key

```bash
# npm
npm install

# pnpm
pnpm install

# yarn
yarn install

# bun
bun install
```

## Development Server

Start the development server on `https://local.beaconcha.in:3000/`:

```bash
# npm
npm run dev

# pnpm
pnpm run dev

# yarn
yarn dev

# bun
bun run dev
```

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


