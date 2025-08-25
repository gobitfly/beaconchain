import type { MessageSchema } from '~/i18n/i18n.config'

export const useTranslation = () => {
  // enables autocompletion
  // https://vue-i18n.intlify.dev/guide/advanced/typescript.html#resource-keys-completion-supporting
  return { ...useI18n<{ message: MessageSchema }>({ useScope: 'global' }) }
}
export type TranslationInput
    = | TranslationKey
      | TranslationWithMessageInterpolation

export type TranslationKey = GetObjectPaths<MessageSchema>

/**
 * @example
 *
 * For cases like:
 * - Plural $t('common.message', 2) -> "Messages"
 * - Named Interpolation $t('common.message', { name: 'World' }) -> "Hello World"
 * - List Interpolation $t('common.message', ['World', 'Beaconchain']) -> "Hello World, this is Beaconchain"
 *
 * @see https://vue-i18n.intlify.dev/api/injection.html#component-injections
 */
type TranslationWithMessageInterpolation
  = | {
    interpolation: number,
    key: TranslationKey,
  }
  | {
    interpolation: Record<string, unknown>,
    key: TranslationKey,
  }
  | {
    interpolation: string[],
    key: TranslationKey,
  }
