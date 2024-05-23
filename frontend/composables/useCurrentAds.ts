import { useAdConfigurationStore } from '~/stores/useAdConfigurationStore'
import type { AdConfiguration } from '~/types/adConfiguration'

export function useCurrentAds () {
  const { adConfigs, refreshAdConfigs } = useAdConfigurationStore()
  const { user } = useUserStore()
  const { path, name } = useRoute()

  const pathName = computed(() => name?.toString?.() || path)

  watch(pathName, (newName) => {
    refreshAdConfigs(newName)
  }, { immediate: true })

  const ads = computed<AdConfiguration[]>(() => {
    if (user.value?.premium_perks.ad_free) {
      return []
    }

    const configs: AdConfiguration[] = adConfigs.value[pathName.value]?.filter(c => c.enabled) ?? []
    adConfigs.value.global?.forEach((config) => {
      if (config.enabled && !configs.find(c => c.jquery_selector === config.jquery_selector && c.insert_mode === config.insert_mode)) {
        configs.push(config)
      }
    })
    return configs
  })

  return { ads }
}
