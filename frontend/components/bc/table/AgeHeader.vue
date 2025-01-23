<script setup lang="ts">
const {
  toggleAgeFormat,
} = useSettingsStore()
</script>

<template>
  <span
    class="age-header"
    @click="toggleAgeFormat"
  >{{
    $t("table.age")
  }}</span>
</template>

<style lang="scss" scoped>
.age-header {
  color: var(--primary-color);
  cursor: pointer;
}
</style>
