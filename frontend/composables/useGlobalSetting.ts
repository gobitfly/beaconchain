import { SettingDefaults, type GlobalSetting } from '~/types/settings'

export function useGlobalSetting<T extends string> (identifier: GlobalSetting) {
  const cookie = useCookie(identifier)
  const config = SettingDefaults[identifier]

  const setting = computed<T | undefined>(() => {
    if (!cookie.value) {
      return
    }
    return config?.parse ? config.parse<T>(cookie.value) : (config.default) as T
  })

  const changeSetting = (value: T) => { cookie.value = config.toString ? config.toString(value) : value }

  return { setting, changeSetting }
}
