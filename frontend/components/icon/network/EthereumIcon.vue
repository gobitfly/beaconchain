<script setup lang="ts">
const props = defineProps({
  backgroundTheme: { type: String, required: true } // must be either 'light' or 'dark'
})

const ColorMap : Record<string, string[]> = {
  light: ['#343434', '#8C8C8C', '#3C3C3B', '#8C8C8C', '#141414', '#393939'],
  dark: ['#BFBFB6', '#D9D8D1', '#BFBFB6', '#D9D8D1', '#A6A59D', '#BFBFB6']
}
</script>

<template>
  <svg width="34" height="57" viewBox="0 0 34 57" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path d="M16.9957 0L16.624 1.29887V38.9884L16.9957 39.3699L33.9901 29.0286L16.9957 0Z" :fill="ColorMap[props.backgroundTheme][0]" />
    <path d="M16.9954 0L0 29.0286L16.9954 39.3699V21.0765V0Z" :fill="ColorMap[props.backgroundTheme][1]" />
    <path d="M16.9958 42.6822L16.7861 42.9451V56.3708L16.9958 56.9999L34.0004 32.3467L16.9958 42.6822Z" :fill="ColorMap[props.backgroundTheme][2]" />
    <path d="M16.9954 56.9999V42.6822L0 32.3467L16.9954 56.9999Z" :fill="ColorMap[props.backgroundTheme][3]" />
    <path d="M16.9951 39.3696L33.9895 29.0283L16.9951 21.0762V39.3696Z" :fill="ColorMap[props.backgroundTheme][4]" />
    <path d="M0 29.0283L16.9954 39.3696V21.0762L0 29.0283Z" :fill="ColorMap[props.backgroundTheme][5]" />
  </svg>
</template>
