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
    '@nuxtjs/i18n',
  ],
  $meta: {
    name: 'base',
  },
  css: [ join(currentDir, './app/assets/css/main.css') ],
  vite: {
    plugins: [ tailwindcss() ],
  },
  icon: {
    mode: 'css',
    cssLayer: 'base',
    size: '1.25rem',
    localApiEndpoint: '/api/bff/_nuxt_icon',
  },
  /* eslint-enable perfectionist/sort-objects */
})
