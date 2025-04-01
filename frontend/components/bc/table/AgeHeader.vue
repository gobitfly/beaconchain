<script setup lang="ts">
const {
  toggleDateTimeFormat,
} = useSettingsStore()
</script>

<template>
  <BcButtonText
    class="age-header"
    @click.stop="toggleDateTimeFormat"
  >
    {{
      $t("table.age")
    }}
  </BcButtonText>
</template>

<style lang="scss" scoped>
.age-header {
  color: var(--primary-color);
  cursor: pointer;
}
</style>
