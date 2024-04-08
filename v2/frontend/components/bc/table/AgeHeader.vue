<script setup lang="ts">
import type { AgeFormat } from '~/types/settings'

const { setting, changeSetting } = useGlobalSetting<AgeFormat>('age-format')

const toggleAgeSetting = () => {
  changeSetting(setting.value === 'absolute' ? 'relative' : 'absolute')
}
</script>
<template>
  <span class="age" @click="toggleAgeSetting">{{ $t('table.age') }}</span>
</template>

<style lang="scss" scoped>
.age{
  color: var(--primary-color);
  cursor: pointer;
}
</style>
