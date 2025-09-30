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
    size?: 'md' | 'xl',
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
    class="flex justify-center items-center bg-linear-to-b rounded-full font-semibold disabled:opacity-40 aria-disabled:opacity-40 aria-disabled:pointer-events-none active:opacity-80 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-300"
    :class="[
      variant === 'primary' && 'from-gray-700 to-gray-900 text-white dark:from-gray-100 dark:to-gray-300 dark:text-black opacity-90 hover:opacity-95',
      variant === 'secondary' && 'from-gray-300 to-gray-200 text-black dark:from-charcoal-600 dark:to-charcoal-700 dark:text-white opacity-90 hover:opacity-95',
      variant === 'branded' && 'from-brand-500 to-brand-700 text-white hover:opacity-90',
      variant === 'quaternary' && 'text-black dark:text-white hover:opacity-95',
      size === 'md' && 'text-sm py-md px-sm',
      size === 'xl' && 'text-md py-xl px-3xl',
      full ? 'w-full' : 'w-fit',
    ]"
  >
    <LazyBaseIcon
      v-if="leadingIcon"
      :name="leadingIcon"
      class=""
    />
    <span class="px-lg">
      <slot />
    </span>
    <LazyBaseIcon
      v-if="trailingIcon"
      :name="trailingIcon"
      class=""
    />
  </component>
</template>
