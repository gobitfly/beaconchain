import en from '~/locales/en.json'

export type MessageSchema = typeof en

export default defineI18nConfig(() => ({
  legacy: false,
  locale: 'en',
  messages: { en },
}))
