<script setup lang="ts">
const {
  is = 'span',
  screenreaderText,
} = defineProps<{
  is?:
    | 'h2'
    | 'span',
  screenreaderText: TranslationInput,
}>()

const { t: $t } = useTranslation()

const translation = computed(() => {
  if (typeof screenreaderText === 'string') return $t(screenreaderText)
  if (typeof screenreaderText.interpolation === 'number') return $t(screenreaderText.key, screenreaderText.interpolation)
  if (Array.isArray(screenreaderText.interpolation)) return $t(screenreaderText.key, screenreaderText.interpolation)
  return $t(screenreaderText.key, screenreaderText.interpolation)
})
</script>

<template>
  <component
    :is
    class="sr-only"
  >
    {{ translation }}
  </component>
</template>
