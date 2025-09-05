<script setup lang="ts">
import type icons from '@iconify-json/tabler/icons.json'

export type IconName = keyof typeof icons['icons']

const { name } = defineProps<{
  name: IconName,
}>()

// https://github.com/nuxt/icon?tab=readme-ov-file#icon-customization
const customize = (content: string) => {
  return content
    .replace(/stroke-width="[^"]*"/g, 'stroke-width="1.5px"')
}
</script>

<template>
  <Icon
    :name="`tabler:${name}`"
    :customize
  />
</template>
