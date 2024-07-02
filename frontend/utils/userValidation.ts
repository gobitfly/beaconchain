import type { ComposerTranslation } from '@nuxtjs/i18n/dist/runtime/composables'

let t: ComposerTranslation

export function setTranslator (translator: ComposerTranslation) {
  t = translator
}

export function validateEmailAddress (address: string, compareAddress?: string) : true | string {
  if (!address) {
    return t('validation.no_email')
  }
  if (address.length > 100 || !REGEXP_VALID_EMAIL.test(address)) {
    return t('validation.invalid_email')
  }
  if (compareAddress && address !== compareAddress) {
    return t('validation.emails_dont_match')
  }
  return true
}

export function validatePassword (password: string, comparePassword?: string) : true | string {
  if (!password) {
    return t('validation.no_password')
  }
  if (password.length < 5 || password.length > 256) {
    return t('validation.invalid_password')
  }
  if (comparePassword && password !== comparePassword) {
    return t('validation.passwords_dont_match')
  }
  return true
}

export function validateAgreement (agreement : boolean) : true | string {
  if (!agreement) {
    return t('validation.not_agreed')
  }
  return true
}
