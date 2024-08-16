import { defineStore } from 'pinia'
import type { AdConfiguration } from '~/types/adConfiguration'
import { API_PATH } from '~/types/customFetch'

const adConfigurationStore = defineStore('ad_configuration_store', () => {
  const data = ref<Record<string, AdConfiguration[]>>({})
  return { data }
})

export function useAdConfigurationStore() {
  const { fetch } = useCustomFetch()
  const { data } = storeToRefs(adConfigurationStore())

  const adConfigs = computed(() => data.value)

  async function refreshAdConfigs(route: string) {
    const keys = [
      'global',
      route,
    ].join(',')
    const res = await fetch<AdConfiguration[]>(
      API_PATH.AD_CONFIGURATIONs,
      undefined,
      { keys },
    )
    const newConfigurations: Record<string, AdConfiguration[]> = {}
    res.forEach((config) => {
      if (!newConfigurations[config.key]) {
        newConfigurations[config.key] = []
      }
      newConfigurations[config.key].push(config)
    })
    data.value = {
      ...adConfigs.value,
      ...newConfigurations,
    }

    return adConfigs.value
  }

  return {
    adConfigs,
    refreshAdConfigs,
  }
}
