<script setup lang="ts">

interface Props {
  title?: string
  question?: string
  noLabel?: string
  yesLabel?: string
}
const { props, dialogRef } = useBcDialog<Props>({ showHeader: false })

const closeDialog = (response: boolean) => {
  dialogRef?.value.close(response)
}
</script>

<template>
  <div class="content">
    <div class="title">
      {{ props?.title }}
    </div>
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
@use '~/assets/css/fonts.scss';

.content {
  display: flex;
  flex-direction: column;

  .title {
    @include fonts.subtitle_text;
    color: var(--primary-color);
    margin-bottom: var(--padding-small);
  }

  .question {
    flex-grow: 1;
    margin: var(--padding) 0;
    @include fonts.subtitle_text;
  }

  .footer {
    display: flex;
    justify-content: center;
    margin-top: var(--padding);
    gap: var(--padding);
  }
}
</style>
