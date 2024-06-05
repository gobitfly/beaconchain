// https://nuxt.com/docs/api/configuration/nuxt-config
// import path from 'path'

import { nodeResolve } from '@rollup/plugin-node-resolve'
import commonjs from '@rollup/plugin-commonjs'
import { gitDescribeSync } from 'git-describe'
import { warn } from 'vue'
let gitVersion = ''

try {
  const info = gitDescribeSync()
  if (info.raw != null) {
    gitVersion = info.raw
  }
  if (gitVersion === '' && info.hash != null) {
    warn('The GitHub tag of the explorer is unknown. Reading the GitHub hash instead.')
    gitVersion = info.hash
  }
} catch (err) {
  warn('The GitHub tag and hash of the explorer cannot be read with git-describe.')
}

export default defineNuxtConfig({
  devtools: { enabled: true },
  devServer: {
    https: {
      key: 'server.key',
      cert: 'server.crt'
    }
  },
  runtimeConfig: {
    public: {
      apiClient: '',
      legacyApiClient: '',
      apiKey: '',
      gitVersion,
      domain: '',
      v1Domain: '',
      logIp: '',
      logFile: '',
      showInDevelopment: '',
      chainIdByDefault: ''
    },
    private: {
      apiServer: '',
      legacyApiServer: ''
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
    vueI18n: './i18n.config.ts'
  },
  routeRules: {
    '/': {
      redirect: '/dashboard'
    }
  },
  nitro: {
    compressPublicAssets: true
  },
  vite: {
    build: {
      rollupOptions: {
        output: {
          manualChunks (id) {
            if (id.includes('node_modules')) {
              return 'vendor'
            }
          },
          format: 'es'
        },
        plugins: [
          nodeResolve(),
          commonjs()
        ]
      },
      minify: true
    }
  },
  postcss: {
    plugins: {
      autoprefixer: {}
    }
  },
  build: {
    transpile: ['echarts', 'zrender', 'tslib', 'resize-detector']
  }
})
