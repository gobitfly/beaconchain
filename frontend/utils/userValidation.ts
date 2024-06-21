import type { ComposerTranslation } from '@nuxtjs/i18n/dist/runtime/composables'

let t: ComposerTranslation

export function setTranslator (translator: ComposerTranslation) {
  t = translator
}

export function validateAddress (value: string) : true|string {
  if (!value) {
    return t('login_and_register.no_email')
  }
  if (value.length > 100 || !REGEXP_VALID_EMAIL.test(value)) {
    return t('login_and_register.invalid_email')
  }
  return true
}

export function validatePassword (value: string) : true|string {
  if (!value) {
    return t('login_and_register.no_password')
  }
  // TODO: ask for a complex password with special characters and son on?
  if (value.length < 5 || value.length > 256) {
    return t('login_and_register.invalid_password')
  }
  return true
}

export function validateAgreement (value : boolean) : true|string {
  if (!value) {
    return t('login_and_register.not_agreed')
  }
  return true
}
