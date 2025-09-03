import { fileURLToPath } from 'node:url'
import {
  dirname, join,
} from 'node:path'
import tailwindcss from '@tailwindcss/vite'

const currentDir = dirname(fileURLToPath(import.meta.url))

export default defineNuxtConfig({
  $meta: {
    name: 'base',
  },
  css: [ join(currentDir, './app/assets/css/main.css') ],
  vite: {
    plugins: [ tailwindcss() ],
  },
})
