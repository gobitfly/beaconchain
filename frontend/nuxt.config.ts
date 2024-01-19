// https://nuxt.com/docs/api/configuration/nuxt-config
import path from 'path'

export default defineNuxtConfig({
  devtools: { enabled: true },
  runtimeConfig: {
    public: {
      apiClientV1: process.env.API_CLIENT_V1
    },
    private: {
      apiServerV1: process.env.API_SERVER_V1
    }
  },
  css: ['~/assets/css/main.scss'],
  modules: [
    '@nuxtjs/i18n',
    '@nuxtjs/eslint-module',
    '@nuxtjs/color-mode',
    ['@pinia/nuxt', {
      storesDirs: ['./stores/**']
    }],
    ['nuxt-primevue', {
      unstyled: true,
      importPT: { from: path.resolve(__dirname, './presets/lara/') } // import and apply preset
    }]
  ],
  typescript: {
    typeCheck: true
  },
  colorMode: {
    preference: 'light', // default value of $colorMode.preference
    fallback: 'light' // fallback value if not system preference found
  },
  i18n: {
    vueI18n: './i18n.config.ts'
  },
  postcss: {
    plugins: {
      tailwindcss: {},
      autoprefixer: {}
    }
  }
})
