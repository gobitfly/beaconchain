<script setup lang="ts">
import { NuxtLink } from '#components'
import type { NuxtLinkProps } from '#app'
import type { IconName } from '~/layers/base/app/components/BaseIcon.vue'

const {
  size = 'md',
  variant = 'primary',
} = defineProps<
  {
    full?: boolean,
    leadingIcon?: IconName,
    size?: 'lg' | 'md' | 'xl',
    trailingIcon?: IconName,
    variant?: 'branded' | 'primary' | 'quaternary' | 'secondary',
  }
  & (
    | { disabled?: boolean, to?: never }
    | { disabled?: never, to: NuxtLinkProps['to'] }
  )
>()
</script>

<template>
  <component
    :is="to ? NuxtLink : 'button'"
    :to
    :disabled
    class="flex items-center justify-center rounded-full bg-linear-to-b font-semibold active:opacity-80 disabled:opacity-40 aria-disabled:pointer-events-none aria-disabled:opacity-40"
    :class="[
      variant === 'primary' && 'from-gray-700 to-gray-900 text-white opacity-90 hover:opacity-95 dark:from-gray-100 dark:to-gray-300 dark:text-black',
      variant === 'secondary' && 'from-gray-300 to-gray-200 text-black opacity-90 hover:opacity-95 dark:from-charcoal-600 dark:to-charcoal-700 dark:text-white',
      variant === 'branded' && 'from-brand-500 to-brand-700 text-white hover:opacity-90',
      variant === 'quaternary' && 'text-black hover:opacity-95 dark:text-white',
      size === 'md' && 'px-sm py-md text-sm',
      size === 'lg' && 'px-2xl py-lg text-md',
      size === 'xl' && 'px-3xl py-xl text-md',
      full ? 'w-full' : 'min-w-fit',
    ]"
  >
    <LazyBaseIcon
      v-if="leadingIcon"
      :name="leadingIcon"
      class=""
    />
    <span
      :class="[
        size === 'xl' && 'px-lg',
        size === 'lg' && 'px-md',
        size === 'md' && 'px-xs',
      ]"
    >
      <slot />
    </span>
    <LazyBaseIcon
      v-if="trailingIcon"
      :name="trailingIcon"
      class=""
    />
  </component>
</template>
