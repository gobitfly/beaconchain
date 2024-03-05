import type { ComposerTranslation } from '@nuxtjs/i18n/dist/runtime/composables'

export function formatMultiPartSpan (t:ComposerTranslation, key: string, classes:(string | undefined)[], options?: any) {
  const parts = classes.map((c, index) => {
    const value = t(`${key}[${index}]`, options, '')
    const classString = c ? `class="${c}"` : ''
    return `<span ${classString}>${value}</span>`
  })

  return `<span>${parts.join(' ')}</span>`
}
