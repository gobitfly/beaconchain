import { defineStore } from 'pinia'
import type { AdConfiguration } from '~/types/adConfiguration'

export const useAdConfigurationStore = defineStore('ad-configuration', () => {
  const configurations = ref< Record<string, AdConfiguration[]>>({})
  const { fetch } = useCustomFetch()

  async function getAds (route:string) {
    const keys = ['global', route].join(',')
    const res = await fetch<AdConfiguration[]>(API_PATH.AD_CONFIGURATIONs, undefined, { keys })
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
