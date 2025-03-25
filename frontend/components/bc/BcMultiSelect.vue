<script lang="ts" setup>

</script>

<template>
  <MultiSelect
    class="bc-multitselect"
  >
    <template #header>
      <slot name="header" />
    </template>
    <template #value>
      <slot name="value" />
    </template>
    <template #dropdownicon>
      <BcIcon
        name="chevron-down"
      />
    </template>
  </MultiSelect>
</template>

<style lang="scss" scoped>
div.bc-multitselect {
  padding-inline: var(--padding);
  gap: var(--padding-small);
}
</style>
