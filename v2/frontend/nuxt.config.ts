// https://nuxt.com/docs/api/configuration/nuxt-config
// import path from 'path'

export default defineNuxtConfig({
  devtools: { enabled: true },
  runtimeConfig: {
    public: {
      apiClient: process.env.API_CLIENT
    },
    private: {
      apiServer: process.env.API_SERVER
    }
  },
  css: ['~/assets/css/main.scss', '~/assets/css/prime.scss'],
  modules: [
    '@nuxtjs/i18n',
    '@nuxtjs/eslint-module',
    '@nuxtjs/color-mode',
    ['@pinia/nuxt', {
      storesDirs: ['./stores/**']
    }],
    ['nuxt-primevue', {
      /* unstyled: true */
    }]
  ],
  typescript: {
    typeCheck: true
  },
  colorMode: {
    preference: 'system', // default value of $colorMode.preference
    fallback: 'light' // fallback value if not system preference found
  },
  i18n: {
    vueI18n: './i18n.config.ts'
  },
  postcss: {
    plugins: {
      autoprefixer: {}
    }
  }
})
