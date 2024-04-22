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
Go to [fontawesome.com/account/general](https://fontawesome.com/account/general), log in and copy the API key from section "Package Token".

In your `.npmrc` file, replace `FA_TOKEN` with the actual key.

Then type:
```bash
cp .env-example .env
```
In file `.env`, write the URLs of the API servers and the secret key to access them (see the V2 frontend wiki page).

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
Add the following env variable (on Ubuntu: in your `~/.profile`) needed to load local mock data:
```bash
export NODE_TLS_REJECT_UNAUTHORIZED=0
```

Restart.

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


