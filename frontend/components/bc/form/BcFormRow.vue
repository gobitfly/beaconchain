<script lang="ts" setup>

</script>

<template>
  <div class="bc-form-row">
    <slot/>
  </div>
</template>

<style lang="scss" scoped>
.bc-form-row {
  display: grid;
  justify-content: space-between;
  align-items: center;
  grid-template-columns: 1fr auto;
  row-gap: var(--padding-small);
  
  grid-template-areas: 
    "label input"
    "error error";
}
</style>