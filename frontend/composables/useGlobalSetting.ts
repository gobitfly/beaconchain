import { SettingDefaults, type GlobalSetting } from '~/types/settings'

export function useGlobalSetting<T extends string> (identifier: GlobalSetting) {
  const cookie = useCookie(identifier)

  const setting = computed<T>(() => (cookie.value || SettingDefaults[identifier]) as T)

  const changeSetting = (value: T) => { cookie.value = value }

  return { setting, changeSetting }
}
