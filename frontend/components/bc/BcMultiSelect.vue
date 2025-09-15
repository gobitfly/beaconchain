<script lang="ts" setup>

</script>

<template>
  <PvMultiSelect
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
  </PvMultiSelect>
</template>

<style lang="scss" scoped>
div.bc-multitselect {
  padding-inline: var(--padding);
  gap: var(--padding-small);
}
</style>
