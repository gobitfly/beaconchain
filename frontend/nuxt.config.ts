// https://nuxt.com/docs/api/configuration/nuxt-config
// import path from 'path'

// TODO: understanding why 'git-describe' returns an emtpy version tag
//       (maybe we need Semver? see https://www.npmjs.com/package/git-describe )

import { gitDescribeSync } from 'git-describe'
import { warn } from 'vue'
let gitVersion = ''

try {
  const info = gitDescribeSync()
  if (info.tag != null) {
    gitVersion = info.tag
  }
} catch (err) {
  warn('The version number of the explorer cannot be read with git-describe')
}
if (gitVersion === '') {
  warn('The version number of the explorer is unknown. "2" will be shown by default.')
  gitVersion = '2'
}

export default defineNuxtConfig({
  devtools: { enabled: true },
  runtimeConfig: {
    public: {
      apiClient: process.env.API_CLIENT,
      gitVersion
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
