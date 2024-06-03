import type { ComposerTranslation } from '@nuxtjs/i18n/dist/runtime/composables'

export function formatMultiPartSpan (t: ComposerTranslation, key: string, classes: (string | undefined)[], options?: any) {
  const parts = classes.map((c, index) => {
    const value = tOf(t, key, index, options)
    if (!value) {
      return undefined
    }
    const classString = c ? `class="${c}"` : ''
    return `<span ${classString}>${value}</span>`
  }).filter(v => !!v)

  return `<span>${parts.join(' ')}</span>`
}

export function tD (t: ComposerTranslation, path : string, options?: any, d:string = '') : string {
  const translation = t(path, 'NOT_FOUND', { ...options, missingWarn: false })
  return (translation === 'NOT_FOUND') ? d : translation
}

export function tOf (t: ComposerTranslation, path : string, index : number, options?: any) : string {
  return tD(t, `${path}[${index}]`, options)
}

export function tAll (t: ComposerTranslation, path : string, options?: any) : string[] {
  const list: string[] = []
  let index = 0
  for (let value = tOf(t, path, index); value; value = tOf(t, path, ++index, options)) {
    list.push(value)
  }
  return list
}
