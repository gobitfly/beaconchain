<script lang="ts" setup>
import type { MessageSchema } from '~/i18n.config'
import type { KeyPaths } from '~/types/helper'

defineProps<{
  boldpath?: string,
  /**
   * The path to the key in the translation file (e.g. en.json)
   */
  keypath: KeyPaths<MessageSchema>,
  linkpath?: KeyPaths<MessageSchema>,
  tag?: keyof HTMLElementTagNameMap,
  /**
   * URL to link to
   *
   * @example
   *
   * Translation key has to be under `${keypath}._link`
   *
   *  // en.json
   * {
   *  "notifications": {
   *   "template": "For further information {_link}"
   *   "_link": "Click here"
   * }
   */
  to?: string,
}>()
</script>

<template>
  <I18nT
    :keypath
    scope="global"
    :tag="tag || 'span'"
  >
    <template #_bold>
      <slot
        v-if="boldpath"
        name="_bold"
      >
        <span class="bold">
          {{ $t(boldpath) }}
        </span>
      </slot>
    </template>
    <template #_link>
      <slot
        v-if="to && linkpath"
        name="_link"
      >
        <BcLink
          class="link"
          target="_blank"
          :to
        >
          {{ $t(linkpath) }}
        </BcLink>
      </slot>
    </template>
  </I18nT>
</template>

<style lang="scss" scoped>
  .bold {
    font-weight: bold;
  }
</style>
