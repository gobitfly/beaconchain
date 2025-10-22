<script lang="ts" setup>
defineProps<{
  boldpath?: TranslationKey,
  /**
   * The path to the key in the translation file (e.g. en.json)
   */
  keypath: TranslationKey,
  linkpath?: TranslationKey,
  listpath?: TranslationKey,
  tag?: keyof HTMLElementTagNameMap,
  /**
   * URL to link to
   *
   * @example
   *
   * Translation key has to be under `${keypath}.link`
   *
   *  // en.json
   * {
   *  "notifications": {
   *   "template": "For further information {link}"
   *   "link": "Click here"
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
    <template #bold>
      <span
        v-if="boldpath"
        class="bc-translation-bold"
      >{{ $t(boldpath) }}</span>
    </template>
    <template #link>
      <slot
        v-if="to && linkpath"
        name="link"
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
    <template #list>
      <slot
        name="list"
        :listpath
      >
        <ul v-if="listpath">
          <li
            v-for="item in $t(listpath).split('\n')"
            :key="item"
          >
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
