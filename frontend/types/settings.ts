export type GlobalSetting = 'age-format' | 'rpl'

export type AgeFormat = 'absolute' | 'relative'

export type CookieValue = null | string | undefined

interface SettingsGetter {
  <T>(value?: string): T,
}

interface SettingsSetter {
  <T>(value?: T): string,
}

type SettingsConfig = {
  default: unknown,
  parseValue?: SettingsGetter,
  valueToString?: SettingsSetter,
}

const parseValueBoolean = (value: string) => value === 'true'
const valueToStringBoolean = (value: boolean) => value ? 'true' : 'false'

export const SettingDefaults: Record<GlobalSetting, SettingsConfig> = {
  'age-format': { default: 'absolute' },
  'rpl': {
    default: true,
    parseValue: parseValueBoolean as SettingsGetter,
    valueToString: valueToStringBoolean as SettingsSetter,
  },
}
