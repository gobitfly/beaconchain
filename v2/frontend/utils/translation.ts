import type { ComposerTranslation } from '@nuxtjs/i18n/dist/runtime/composables'

export function formatMultiPartSpan (t: ComposerTranslation, key: string, classes: (string | undefined)[], options?: any) {
  const parts = classes.map((c, index) => {
    const value = t(`${key}[${index}]`, options, 'NOT_FOUND')
    if (!value || value === 'NOT_FOUND') {
      return undefined
    }
    const classString = c ? `class="${c}"` : ''
    return `<span ${classString}>${value}</span>`
  }).filter(v => !!v)

  return `<span>${parts.join(' ')}</span>`
}

export function tOf (t: ComposerTranslation, path : string, index : number, options?: any) : string {
  const translation = t(`${path}[${index}]`, options, 'NOT_FOUND')
  return (translation === 'NOT_FOUND') ? '' : translation
}
