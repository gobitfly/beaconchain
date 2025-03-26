<script setup lang="ts">
const {
  toggleAgeFormat,
} = useSettingsStore()
</script>

<template>
  <BcButtonText
    class="age-header"
    @click.stop="toggleAgeFormat"
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
