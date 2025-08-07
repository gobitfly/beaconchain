<script setup lang="ts">
const emit = defineEmits<{
  (e: 'add-validator'): void,
}>()
</script>

<template>
  <BcButtonText
    class="empty delayed-fadein-animation"
    @click="emit('add-validator')"
  >
    {{
      $t("dashboard.validator.summary.add_validator")
    }}
    <BcIcon name="circle-plus" />
  </BcButtonText>
</template>

<style lang="scss" scoped>
.empty {
  font-family: var(--big_text_font_family);
  font-size: var(--big_text_font_size);
  font-weight: var(--big_text_font_weight);
  width: 100%;
  height: 400px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  color: var(--text-color-disabled);
  gap: var(--padding);
  cursor: pointer;

  svg {
    width: 30px;
    height: 30px;
  }
}
</style>
