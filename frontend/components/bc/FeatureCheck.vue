<script lang="ts" setup>
defineProps<{
  available?: boolean,
}>()
</script>

<template>
  <BcIcon
    v-if="available"
    name="check"
    class="checkmark"
  />
  <BcIcon
    v-else
    name="xmark"
    class="cross"
  />
</template>

<style lang="scss" scoped>
.checkmark {
  color: var(--primary-orange);
  width: 18px;
  height: 20px;
}

.cross {
  color: var(--text-color-discreet);
  width: 14px;
  height: 20px;
}
</style>
