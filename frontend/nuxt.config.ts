// https://nuxt.com/docs/api/configuration/nuxt-config
// import path from 'path'

import { gitDescribeSync } from 'git-describe'
import { warn } from 'vue'
let gitVersion = ''

try {
  const info = gitDescribeSync()
  if (info.tag != null) {
    gitVersion = info.tag
  }
  if (gitVersion === '' && info.hash != null) {
    warn('The GitHub tag of the explorer is unknown. Reading the GitHub hash instead.')
    gitVersion = info.hash
  }
} catch (err) {
  warn('The GitHub tag and hash of the explorer cannot be read with git-describe.')
}
if (gitVersion === '') {
  warn('Neither the GitHub tag nor the GitHub hash could be read. The version number of the explorer shown on the front end will be "2" by default.')
  gitVersion = '2'
}

export default defineNuxtConfig({
  devtools: { enabled: true },
  runtimeConfig: {
    public: {
      apiClient: process.env.API_CLIENT,
      legacyApiClient: process.env.LEGACY_API_CLIENT,
      gitVersion
    },
    private: {
      apiServer: process.env.API_SERVER,
      legacyApiServer: process.env.LEGACY_API_SERVER
    }
  },
  css: ['~/assets/css/main.scss', '~/assets/css/prime.scss', '@fortawesome/fontawesome-svg-core/styles.css'],
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
    fallback: 'dark' // fallback value if not system preference found
  },
  i18n: {
    vueI18n: './i18n.config.ts',
    compilation: {
      strictMessage: false
    }
  },
  postcss: {
    plugins: {
      autoprefixer: {}
    }
  }
})
