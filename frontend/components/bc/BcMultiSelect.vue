<script lang="ts" setup>

</script>

<template>
  <MultiSelect>
    <template #header>
      <slot name="header" />
    </template>
    <template #value>
      <slot name="value" />
    </template>
    <template #dropdownicon>
      <IconChevron direction="bottom" />
    </template>
  </MultiSelect>
</template>
