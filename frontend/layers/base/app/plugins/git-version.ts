declare global {
  interface Window {
    gitVersion: string,
  }
}

export default defineNuxtPlugin({
  hooks: {
    'app:mounted'() {
      window.gitVersion = useRuntimeConfig().public.gitVersion
    },
  },
  name: 'git-version',
  setup() {
    const { gitVersion } = useRuntimeConfig().public
    useHead({
      htmlAttrs: {
        'data-git-version': gitVersion,
      },
    })
  },
})
