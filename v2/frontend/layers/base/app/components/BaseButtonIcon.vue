<script setup lang="ts">
import type { NuxtLinkProps } from '#app'
import { NuxtLink } from '#components'
// https://github.com/vuejs/core/issues/8952
import type { IconName } from '~/layers/base/app/components/BaseIcon.vue'

const {
  size = 'md',
} = defineProps<(
  {
    // eslint-disable-next-line vue/prop-name-casing -- conditional props for props like `ariaLabel` do not work
    'aria-labelledby': string,
    'screenreaderText'?: never,
  }
  | {
    // eslint-disable-next-line vue/prop-name-casing -- conditional props for props like `ariaLabel` do not work
    'aria-labelledby'?: never,
    'screenreaderText': TranslationInput,
  }
)
& {
  name: IconName,
  size?: 'lg' | 'md',
  to?: NuxtLinkProps['to'],
  variant: 'secondary' | 'tertiary',
}
>()
</script>

<template>
  <component
    :is="to ? NuxtLink : 'button'"
    :to
    class="border flex rounded-full bg-linear-to-b disabled:opacity-40 aria-disabled:opacity-40 active:opacity-80 size-fit"
    :class="[
      variant === 'secondary' && 'from-gray-300 to-gray-200 dark:from-charcoal-600 dark:to-charcoal-700 opacity-90 border-transparent',
      variant === 'tertiary' && 'dark:border-charcoal-400 border-gray-200',
      size === 'md' && 'p-md',
      size === 'lg' && 'p-md',
    ]"
  >
    <LazyBaseScreenreaderOnly
      v-if="screenreaderText"
      :screenreader-text
    />
    <BaseIcon
      :name
      :class="[
        size === 'lg' && 'text-4xl',
      ]"
    />
  </component>
</template>
