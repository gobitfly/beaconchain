import { warn } from 'vue'

export default defineNuxtPlugin((_nuxtApp) => {
  return {
    provide: {
      bcLogger: {
        warn: (...rest:any) => warn('Warn', rest)
      }
    }
  }
})
