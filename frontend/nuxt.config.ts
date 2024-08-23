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
    warn(
      'The GitHub tag of the explorer is unknown. Reading the GitHub hash instead.',
    )
    gitVersion = info.hash
  }
}
catch (err) {
  warn(
    'The GitHub tag and hash of the explorer cannot be read with git-describe.',
  )
}

export default defineNuxtConfig({
  build: {
    transpile: [
      'echarts',
      'zrender',
      'tslib',
      'resize-detector',
    ],
  },
  colorMode: {
    fallback: 'dark', // fallback value if not system preference found
    preference: 'system', // default value of $colorMode.preference
  },
  compatibilityDate: '2024-07-15',
  css: [
    '~/assets/css/main.scss',
    '~/assets/css/prime.scss',
    '@fortawesome/fontawesome-svg-core/styles.css',
  ],
  devServer: {
    host: 'local.beaconcha.in',
    https: true,
  },
  devtools: { enabled: true },
  eslint: { config: { stylistic: true } },
  i18n: { vueI18n: './i18n.config.ts' },
  modules: [
    '@nuxtjs/i18n',
    '@nuxtjs/color-mode',
    [
      '@pinia/nuxt',
      { storesDirs: [ './stores/**' ] },
    ],
    '@primevue/nuxt-module',
    '@nuxt/eslint',
  ],
  nitro: { compressPublicAssets: true },
  postcss: { plugins: { autoprefixer: {} } },
  routeRules: { '/': { redirect: '/dashboard' } },
  runtimeConfig: {
    private: {
      apiServer: process.env.PRIVATE_API_SERVER,
      legacyApiServer: process.env.PRIVATE_LEGACY_API_SERVER,
      ssrSecret: process.env.PRIVATE_SSR_SECRET || '',
    },
    public: {
      apiClient: process.env.PUBLIC_API_CLIENT,
      apiKey: process.env.PUBLIC_API_KEY,
      chainIdByDefault: process.env.PUBLIC_CHAIN_ID_BY_DEFAULT,
      domain: process.env.PUBLIC_DOMAIN,
      gitVersion,
      legacyApiClient: process.env.PUBLIC_LEGACY_API_CLIENT,
      logFile: '',
      logIp: '',
      maintenanceTS: '',
      showInDevelopment: '',
      stripeBaseUrl: process.env.PUBLIC_STRIPE_BASE_URL,
      v1Domain: process.env.PUBLIC_V1_DOMAIN,
    },
  },
  ssr: process.env.ENABLE_SSR !== 'FALSE',
  vite: {
    build: {
      minify: true,
      rollupOptions: {
        output: {
          format: 'es',
          manualChunks(id) {
            if (id.includes('node_modules')) {
              return 'vendor'
            }
          },
        },
        plugins: [
          nodeResolve(),
          commonjs(),
        ],
      },
    },
    esbuild: {
      drop: [ 'console' ],
    },
  },
})
