import * as fs from 'fs'
import * as path from 'path'
import { warn } from 'vue'

export default defineNuxtPlugin((_nuxtApp) => {
  const { public: { logFile } } = useRuntimeConfig()
  return {
    provide: {
      bcLogger: {
        warn: (msg: string, ...rest:any) => {
          const ts = new Date().toISOString()
          if (process.server && logFile) {
            const filePath = path.resolve(logFile)
            fs.appendFileSync(filePath, `${ts}: ${msg} | ${JSON.stringify(rest)}\n`)
          }
          warn(`${ts}: `, ...rest)
        }
      }
    }
  }
})
