<script setup lang="ts">
import type { NuxtLinkProps } from '#app'
import { NuxtLink } from '#components'
// https://github.com/vuejs/core/issues/8952
import type { IconName } from '~/layers/base/app/components/BaseIcon.vue'

const {
  size = 'md',
  to,
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
  isDisabled?: boolean,
  name: IconName,
  size?: 'lg' | 'md',
  to?: NuxtLinkProps['to'],
  variant: 'secondary' | 'tertiary',
}
>()
const isButton = computed(() => !to)
</script>

<template>
  <component
    :is="isButton ? 'button' : NuxtLink"
    :type="isButton ? 'button' : undefined"
    :disabled="isDisabled"
    :to
    class="flex size-fit rounded-full border bg-linear-to-b active:opacity-80 disabled:opacity-40 aria-disabled:opacity-40"
    :class="[
      variant === 'secondary' && 'border-transparent from-gray-300 to-gray-200 opacity-90 dark:from-charcoal-600 dark:to-charcoal-700',
      variant === 'tertiary' && 'border-gray-200 dark:border-charcoal-400',
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
