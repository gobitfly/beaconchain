import en from '~/i18n/locales/en.json'

export type MessageSchema = typeof en
const messages = { 'en-US': en }
export type Locale = keyof typeof messages

export default defineI18nConfig(() => ({
  legacy: false,
  locale: 'en-US',
  messages,
}))
