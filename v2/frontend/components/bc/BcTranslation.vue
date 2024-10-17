<script lang="ts" setup>
import type { MessageSchema } from '~/i18n.config'
import type { KeyPaths } from '~/types/helper'

defineProps<{
  boldpath?: KeyPaths<MessageSchema>,
  /**
   * The path to the key in the translation file (e.g. en.json)
   */
  keypath: KeyPaths<MessageSchema>,
  linkpath?: KeyPaths<MessageSchema>,
  listpath?: KeyPaths<MessageSchema>,
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
      <span v-if="boldpath" class="bc-translation-bold">{{ $t(boldpath) }}</span>
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
    <template #_list>
      <slot name="_list" :listpath>
        <ul v-if="listpath">
          <li v-for="item in $t(listpath).split('\n')" :key="item">
            {{ item }}
          </li>
        </ul>
      </slot>
    </template>
  </I18nT>
</template>

<style lang="scss" scoped>
.bc-translation-bold {
  font-weight: 800;
}
</style>
