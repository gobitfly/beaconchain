export type GlobalSetting = 'age-format'

export type AgeFormat = 'absolut' | 'relative'

interface SettingsGetter {
  <T>(value?: string): T;
}

interface SettingsSetter {
  <T>(value?: T): string;
}

type SettingsConfig = {
  default: unknown,
  parse?: SettingsGetter,
  toString?: SettingsSetter
}

export const SettingDefaults:Record<GlobalSetting, SettingsConfig> = {
  'age-format': {
    default: 'absolute'
  }
}
