import { SettingDefaults, type GlobalSetting } from '~/types/settings'

export function useGlobalSetting<T extends string> (identifier: GlobalSetting) {
  const cookie = useCookie(identifier)
  const config = SettingDefaults[identifier]

  const setting = computed<T | undefined>(() => {
    if (!cookie.value) {
      return config.default as T
    }
    return config?.parseValue ? config.parseValue<T>(cookie.value) : cookie.value as T
  })
  const changeSetting = (value: T) => {
    cookie.value = config.valueToString !== undefined ? config.valueToString(value) : value
  }

  return { setting, changeSetting }
}
