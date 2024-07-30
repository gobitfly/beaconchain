import type { ComposerTranslation } from '@nuxtjs/i18n/dist/runtime/composables'

const NOT_FOUND = 'NOT_FOUND'
const END_OF_ARRAY = 'END_OF_ARRAY'

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

/** Translation with default. Needed if we want to default to '' as the translation lib does not support to default to '' */
export function tD (t: ComposerTranslation, path : string, options?: any, d:string = '') : string {
  if (typeof options === 'number') {
    options = { plural: options }
  }
  const translation = t(path, NOT_FOUND, { ...options, missingWarn: false })
  return (translation === NOT_FOUND) ? d : translation
}

function tOfWithEOA (t: ComposerTranslation, path : string, index : number, options?: any) : string {
  return tD(t, `${path}[${index}]`, options, END_OF_ARRAY)
}

/** gets a translation at a specific index, if it exists */
export function tOf (t: ComposerTranslation, path : string, index : number, options?: any) : string {
  const element = tOfWithEOA(t, path, index, options)
  return (element === END_OF_ARRAY) ? '' : element
}

/** returns an array of translations (with only one element if the translation data is a simple string) */
export function tAll (t: ComposerTranslation, path : string, options?: any) : string[] {
  const list: string[] = []
  if (tD(t, path, options, NOT_FOUND) !== NOT_FOUND) {
    // the data is a string, we push the translation
    list.push(tD(t, path, options))
  } else {
    // the data is an array
    let index = 0
    for (let value = tOfWithEOA(t, path, index, options); value !== END_OF_ARRAY; value = tOfWithEOA(t, path, ++index, options)) {
      list.push(value)
    }
  }
  return list
}

/** returns an array of translations path's for a list */
export function hasTranslation (t: ComposerTranslation, path : string) {
  if (tD(t, path, undefined, NOT_FOUND) !== NOT_FOUND) {
    return true
  }
  return false
}
