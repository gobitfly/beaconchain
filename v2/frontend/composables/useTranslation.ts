import type { MessageSchema } from '~/i18n.config'

export function useTranslation() {
  // enables autocompletion
  // https://vue-i18n.intlify.dev/guide/advanced/typescript.html#resource-keys-completion-supporting
  return { ...useI18n<{ message: MessageSchema }>({ useScope: 'global' }) }
}
