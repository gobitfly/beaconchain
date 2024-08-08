import type { ComposerTranslation } from 'vue-i18n'
import type { BooleanSchema, StringSchema } from 'yup'
import { boolean as yupBool, ref as yupRef, string as yupString } from 'yup'

export function passwordValidation(t: ComposerTranslation): StringSchema {
  return yupString()
    .required(t('validation.password.empty'))
    .min(5, t('validation.password.min', { amount: 5 }))
    .max(64, t('validation.password.max', { amount: 64 }))
}

export function confirmPasswordValidation(
  t: ComposerTranslation,
  comparerRefName: string,
): StringSchema {
  return passwordValidation(t).oneOf(
    [yupRef(comparerRefName)],
    t('validation.password.no_match'),
  )
}

export function newPasswordValidation(
  t: ComposerTranslation,
  oldRefName: string,
): StringSchema {
  return passwordValidation(t).notOneOf(
    [yupRef(oldRefName)],
    t('validation.password.not_new'),
  )
}

export function emailValidation(t: ComposerTranslation): StringSchema {
  return yupString()
    .required(t('validation.email.empty'))
    .matches(REGEXP_VALID_EMAIL, t('validation.email.invalid'))
}

export function confirmEmailValidation(
  t: ComposerTranslation,
  comparerRefName: string,
): StringSchema {
  return emailValidation(t).oneOf(
    [yupRef(comparerRefName)],
    t('validation.email.no_match'),
  )
}

export function checkboxValidation(errorMessage: string): BooleanSchema {
  return yupBool().test('is-true', errorMessage, value => value === true)
}
