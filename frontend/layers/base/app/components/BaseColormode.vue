<script setup lang="ts">
const colorMode = useColorMode()
const cookie = useBcCookie<'dark' | 'light'>('theme')
colorMode.preference = cookie.value || 'dark'
const handleUpdate = () => {
  cookie.value = colorMode.preference as 'dark' | 'light'
  // due to a bug in @nuxtjs/color-mode
  // `data-theme` does not update when user comes back to the app
  // using browser back/forward buttons
  document.documentElement.dataset.theme = colorMode.preference
}
</script>

<template>
  <BaseSwitch
    v-model="colorMode.preference"
    class="w-fit"
    :values="[
      {
        label: $t('base.footer.color_mode.light'),
        key: 'light',
      },
      {
        label: $t('base.footer.color_mode.dark'),
        key: 'dark',
      },
    ]"
    :class-list="{
      trackItem: 'p-md +py-md +px-lg border border-transparent rounded-4xl text-gray-600 dark:text-gray-400 has-checked:text-black has-checked:dark:text-white has-focus-visible:outline-2 has-focus-visible:outline-offset-2 has-focus-visible:outline-brand-300',
      track: 'border border-gray-100 dark:border-gray-900 flex gap-md [&>*]:grow text-center bg-white dark:bg-gray-950 rounded-4xl shadow-[0_1px_0.5px_0_rgba(255,255,255,0.08)_inset,0-1px_0_0_rgba(255,255,255,0.18)_inset]',
      thumb: 'bg-gray-50 dark:bg-gray-800 rounded-4xl shadow-[0_2px_2px_0_rgba(0,0,0,0.25),_0_0.5px_0.5px_0_rgba(255,255,255,0.12)_inset]',
    }"
    screenreader-title="base.footer.color_mode.title"
    @update:model-value="handleUpdate"
  >
    <template #light=" { label } ">
      <BaseIcon name="sun" />
      <span class="sr-only">{{ label }}</span>
    </template>
    <template #dark=" { label } ">
      <BaseIcon name="moon" />
      <span class="sr-only">{{ label }}</span>
    </template>
  </BaseSwitch>
</template>

<style scoped></style>
