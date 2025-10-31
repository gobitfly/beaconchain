<script lang="ts" setup>
/**
 * This component provides two approaches for translation interpolation:
 * 1. Props (simple patterns): suppath, linkpath, boldpath, listpath
 * 2. Named slots (complex/dynamic content): for computed values or multiple interpolations
 *
 * @example Simple - using props
 * <BcTranslation
 *   keypath="base.common.beaconscore"
 *   suppath="base.common.registered_trademark_symbol"
 * />
 *
 * @example Complex - using named slots
 * <BcTranslation keypath="dashboard.validator.summary.tooltip.higher">
 *   <template #sup><sup>®</sup></template>
 *   <template #name>{{ groupName }}</template>
 *   <template #average>{{ formatPercent(value) }}</template>
 * </BcTranslation>
 */
defineProps<{
  /** Translation key for bold text content */
  boldpath?: TranslationKey,
  /** Translation key (required) */
  keypath: TranslationKey,
  /** Translation key for link text (requires `to` prop) */
  linkpath?: TranslationKey,
  /** Translation key for newline-separated list items */
  listpath?: TranslationKey,
  /** Translation key for superscript content (e.g., trademark symbols) */
  suppath?: TranslationKey,
  /** HTML tag to wrap translation (default: 'span') */
  tag?: keyof HTMLElementTagNameMap,
  /** URL for link (requires `linkpath` prop) */
  to?: string,
}>()

const slots = useSlots()
</script>

<template>
  <I18nT
    :keypath
    scope="global"
    :tag="tag || 'span'"
  >
    <!-- Forward custom named slots (use only for dynamic/complex content) -->
    <template
      v-for="(_, name) in slots"
      :key="String(name)"
      #[name]
    >
      <slot
        :name="String(name)"
      />
    </template>

    <!-- Prop-based templates (preferred - use these via props when possible) -->
    <template
      v-if="suppath && !slots.sup"
      #sup
    >
      <sup>{{ $t(suppath) }}</sup>
    </template>

    <template
      v-if="boldpath && !slots.bold"
      #bold
    >
      <span class="bc-translation-bold">{{ $t(boldpath) }}</span>
    </template>

    <template
      v-if="linkpath && to && !slots.link"
      #link
    >
      <BcLink
        :to
        class="link"
        target="_blank"
      >
        {{ $t(linkpath) }}
      </BcLink>
    </template>

    <template
      v-if="listpath && !slots.list"
      #list
    >
      <ul>
        <li
          v-for="item in $t(listpath).split('\n')"
          :key="item"
        >
          {{ item }}
        </li>
      </ul>
    </template>
  </I18nT>
</template>

<style lang="scss" scoped>
.bc-translation-bold {
  font-weight: 800;
}
</style>
