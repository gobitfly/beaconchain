import { fileURLToPath } from 'node:url'
import {
  dirname, join,
} from 'node:path'
import tailwindcss from '@tailwindcss/vite'

const currentDir = dirname(fileURLToPath(import.meta.url))

export default defineNuxtConfig({
  /* eslint-disable perfectionist/sort-objects  -- as there is a conflict with `nuxt specific eslint rules` */
  modules: [
    '@nuxt/icon',
    '@nuxtjs/color-mode',
    '@nuxtjs/i18n',
    'reka-ui/nuxt',
  ],
  $meta: {
    name: 'base',
  },
  css: [ join(currentDir, './app/assets/css/main.css') ],
  router: {
    options: {
      scrollBehaviorType: 'smooth',
    },
  },
  colorMode: {
    fallback: 'dark',
    preference: 'dark',
    dataValue: 'theme',
    storageKey: 'theme',
    // currently cookie storage is only applying the theme on the current path
    // See open PR: https://github.com/nuxt-modules/color-mode/pull/301
    // storage: 'cookie',
    storage: 'localStorage',
  },
  vite: {
    plugins: [ tailwindcss() ],
  },
  icon: {
    mode: 'css',
    cssLayer: 'base',
    localApiEndpoint: '/api/bff/_nuxt_icon',
  },
  reka: {
    prefix: 'Rk',
  },
  /* eslint-enable perfectionist/sort-objects */
})
