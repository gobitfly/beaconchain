import type { ComposerTranslation } from '@nuxtjs/i18n/dist/runtime/composables'

export function formatMultiPartSpan (t: ComposerTranslation, key: string, classes:(string | undefined)[], links?:(string | undefined)[], options?: any) {
  const parts = classes.map((c, index) => {
    const value = t(`${key}[${index}]`, options, '')
    if (!value) {
      return undefined
    }
    const classString = c ? `class="${c}"` : ''

    const spanElement = `<span ${classString}>${value}</span>`

    if (links?.[index]) {
      return `<a href="${links[index]}" target="_blank">${spanElement}</a>`
    }

    return spanElement
  }).filter(v => !!v)

  return `<span>${parts.join(' ')}</span>`
}

export function tOf (t: ComposerTranslation, path : string, index : number) : string {
  return t(`${path}[${index}]`)
}
