export type AgeFormat = 'absolute' | 'relative'

export type CookieValue = null | string | undefined

export type GlobalSetting = 'age-format' | 'rpl'

type SettingsConfig = {
  default: unknown,
  parseValue?: SettingsGetter,
  valueToString?: SettingsSetter,
}

interface SettingsGetter {
  <T>(value?: string): T,
}

interface SettingsSetter {
  <T>(value?: T): string,
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
