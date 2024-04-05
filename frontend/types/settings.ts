export type GlobalSetting = 'age-format'

export type AgeFormat = 'absolut' | 'relative'

export const SettingDefaults:Record<GlobalSetting, unknown> = {
  'age-format': 'absolut'
}
