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
Checkout the `beaconchain` repository from git and navigate to `beaconchain/frontend`.

Type
```
cp .npmrc-example .npmrc
```
Go to [fontawesome.com/account/general](https://fontawesome.com/account/general), log in and copy the API key from section "Package Token" in Bitwarden.

In your `.npmrc` file, replace `YOURKEY` with the actual key.

Then type:
```
cp .env-example .env
```
Write the following settings in file `.env`:
```
NUXT_PUBLIC_API_CLIENT: "https://holesky.beaconcha.in/api/i/"
NUXT_PUBLIC_LEGACY_API_CLIENT: "https://sepolia.beaconcha.in/"
NUXT_PUBLIC_API_KEY: "the 'V2 API Key Secret' stored by Bitwarden in the Employee folder"
NUXT_PRIVATE_API_SERVER: "https://holesky.beaconcha.in/api/i/"
NUXT_PRIVATE_LEGACY_API_SERVER: "https://sepolia.beaconcha.in/"

```

Set the following mapping in your `/etc/hosts` file:
```
127.0.0.1 local.beaconcha.in
```

Create server certificates for locally running on https, by runing these comands in the console (the last two with `sudo`)
```bash
openssl genrsa 2048 > server.key
chmod 400 server.key
openssl req -new -x509 -nodes -sha256 -days 365 -key server.key -out server.crt
```
Set the following env variable (needed to load local mock data): 
`export NODE_TLS_REJECT_UNAUTHORIZED=0`

Restart.

Run
```bash
npm install
sudo npm install -g pnpm
sudo npm install -g yarn
sudo npm install -g bun
```
and then
```bash
pnpm install
yarn install
bun install
```

## Development Server

Start the development server on `https://local.beaconcha.in:3000/`:

```bash
npm run dev & pnpm run dev & yarn dev & bun run dev
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


