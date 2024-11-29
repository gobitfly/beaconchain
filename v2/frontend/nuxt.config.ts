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
catch {
  warn(
    'The GitHub tag and hash of the explorer cannot be read with git-describe.',
  )
}

export default defineNuxtConfig({
  /* eslint-disable perfectionist/sort-objects  -- as there is a conflict with `nuxt specific eslint rules` */
  modules: [
    '@nuxtjs/i18n',
    '@nuxtjs/color-mode',
    [
      '@pinia/nuxt',
      { storesDirs: [ './stores/**' ] },
    ],
    '@primevue/nuxt-module',
    '@nuxt/eslint',
    '@vueuse/nuxt',
  ],
  ssr: process.env.ENABLE_SSR !== 'FALSE',
  devtools: { enabled: true },
  css: [
    '~/assets/css/main.scss',
    '~/assets/css/prime.scss',
    '@fortawesome/fontawesome-svg-core/styles.css',
  ],
  colorMode: {
    fallback: 'dark',
    preference: 'dark',
  },
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
      deploymentType: process.env.PUBLIC_DEPLOYMENT_TYPE,
      domain: process.env.PUBLIC_DOMAIN,
      gitVersion,
      isApiMocked: '',
      legacyApiClient: process.env.PUBLIC_LEGACY_API_CLIENT,
      logFile: '',
      maintenanceTS: '',
      showInDevelopment: '',
      stripeBaseUrl: process.env.PUBLIC_STRIPE_BASE_URL,
      v1Domain: process.env.PUBLIC_V1_DOMAIN,
    },
  },
  build: {
    transpile: [
      'echarts',
      'zrender',
      'tslib',
      'resize-detector',
    ],
  },
  routeRules: { '/': { redirect: '/dashboard' } },
  devServer: {
    host: 'local.beaconcha.in',
    https: {
      cert: 'server.crt',
      key: 'server.key',
    },
  },
  compatibilityDate: '2024-07-15',
  nitro: {
    compressPublicAssets: true,
    esbuild: {
      options: {
        target: 'esnext',
      },
    },
  },
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
    css: {
      preprocessorOptions: {
        scss: {
          api: 'modern-compiler',
        },
      },
    },
  },
  postcss: { plugins: { autoprefixer: {} } },
  eslint: { config: { stylistic: true } },
  i18n: { vueI18n: './i18n.config.ts' },
  /* eslint-enable perfectionist/sort-objects */
})
