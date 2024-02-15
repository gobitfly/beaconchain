import { storeToRefs } from 'pinia'
import { useAdConfigurationStore } from '~/stores/useAdConfigurationStore'
import type { AdConfiguration } from '~/types/adConfiguration'

export function useCurrentAds () {
  const { getAds } = useAdConfigurationStore()
  const { configurations } = storeToRefs(useAdConfigurationStore())
  const { path } = useRoute()

  watch(() => path, (newPath) => {
    getAds(newPath)
  }, { immediate: true })

  // TODO: also validate if user is premium user and config is not for all
  const ads = computed(() => {
    const configs: AdConfiguration[] = configurations.value[path]?.filter(c => c.enabled) ?? []
    configurations.value.global?.forEach((config) => {
      if (config.enabled && !configs.find(c => c.jquery_selector === config.jquery_selector && c.insert_mode === config.insert_mode)) {
        configs.push(config)
      }
    })
    return configs
  })

  return { ads }
}
