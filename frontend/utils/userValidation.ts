import type { ComposerTranslation } from '@nuxtjs/i18n/dist/runtime/composables'
import { ref as yupRef, string as yupString, StringSchema } from 'yup'

let t: ComposerTranslation

export function setTranslator (translator: ComposerTranslation) {
  t = translator
}

export function validateAgreement (agreement : boolean) : true | string {
  if (!agreement) {
    return t('validation.not_agreed')
  }
  return true
}

export function passwordValidation (t: ComposerTranslation) : StringSchema {
  return yupString().required(t('validation.password.empty')).min(5, t('validation.password.min', { amount: 5 })).max(64, t('validation.password.max', { amount: 64 }))
}

export function confirmPasswordValidation (t: ComposerTranslation, comparerRefName: string) : StringSchema {
  return passwordValidation(t).oneOf([yupRef(comparerRefName)], t('validation.password.no_match'))
}

export function emailValidation (t: ComposerTranslation) : StringSchema {
  return yupString().required(t('validation.email.empty')).email(t('validation.email.invalid'))
}

export function confirmEmailValidation (t: ComposerTranslation, comparerRefName: string) : StringSchema {
  return emailValidation(t).oneOf([yupRef(comparerRefName)], t('validation.email.no_match'))
}
