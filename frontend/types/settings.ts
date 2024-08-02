export type GlobalSetting = 'age-format' | 'rpl'

export type AgeFormat = 'absolute' | 'relative'

interface SettingsGetter {
  <T>(value?: string): T;
}

interface SettingsSetter {
  <T>(value?: T): string;
}

type SettingsConfig = {
  default: unknown,
  parseValue?: SettingsGetter,
  valueToString?: SettingsSetter
}

export const SettingDefaults:Record<GlobalSetting, SettingsConfig> = {
  'age-format': {
    default: 'absolute'
  },
  rpl: {
    default: true
  }
}
