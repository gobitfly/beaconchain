<script setup lang="ts">

interface Props {
  question?: string
  noLabel?: string
  yesLabel?: string
}
const { props, dialogRef } = useBcDialog<Props>()

const closeDialog = (response: boolean) => {
  dialogRef?.value.close(response)
}
</script>

<template>
  <div class="content">
    <div class="question">
      {{ props?.question }}
    </div>
    <div class="footer">
      <Button type="button" :label="props?.noLabel || $t('navigation.no')" @click="closeDialog(false)" />
      <Button type="button" :label="props?.yesLabel || $t('navigation.yes')" @click="closeDialog(true)" />
    </div>
  </div>
</template>

<style lang="scss" scoped>
.content {
  display: flex;
  flex-direction: column;

  .question {
    flex-grow: 1;
    margin: var(--padding) 0;
  }

  .footer {
    display: flex;
    justify-content: flex-end;
    gap: var(--padding);
  }
}
</style>
