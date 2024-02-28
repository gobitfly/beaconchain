import { defineStore } from 'pinia'
import type { AdConfiguration } from '~/types/adConfiguration'

export const useAdConfigurationStore = defineStore('ad-configuration', () => {
  const configurations = ref< Record<string, AdConfiguration[]>>({})

  async function getAds (route:string) {
    const keys = ['global', route].join(',')
    const res = await useCustomFetch<AdConfiguration[]>(API_PATH.AD_CONFIGURATIONs, undefined, { keys })
    const newConfigurations:Record<string, AdConfiguration[]> = {}
    res.forEach((config) => {
      if (!newConfigurations[config.key]) {
        newConfigurations[config.key] = []
      }
      newConfigurations[config.key].push(config)
    })
    configurations.value = { ...configurations.value, ...newConfigurations }
    return configurations.value
  }

  return { configurations, getAds }
})
