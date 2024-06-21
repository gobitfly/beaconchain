<script setup lang="ts">

interface Props {
  title?: string
  warning?: string
  noLabel?: string // defaults to "Cancel"
  yesLabel?: string // defaults to "Delete"
}
const { props, dialogRef } = useBcDialog<Props>({ showHeader: false })

const closeDialog = (response: boolean) => {
  dialogRef?.value.close(response)
}
</script>

<template>
  <div class="content">
    <div v-if="props?.title" class="title">
      {{ props?.title }}
    </div>
    <div v-if="props?.warning" class="warning">
      {{ props?.warning }}
    </div>
    <div class="footer">
      <div class="cancel" @click="closeDialog(false)">
        {{ props?.noLabel || $t('navigation.cancel') }}
      </div>
      <Button type="button" class="delete" :label="props?.yesLabel || $t('navigation.delete')" @click="closeDialog(true)" />
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use '~/assets/css/main.scss';
@use '~/assets/css/fonts.scss';

.content {
  display: flex;
  flex-direction: column;

  .title {
    @include fonts.subtitle_text;
    color: var(--primary-color);
    margin-bottom: var(--padding-small);
  }

  .warning {
    flex-grow: 1;
    @include fonts.tiny_text;
    font-weight: var(--roboto-medium);
    margin: var(--padding) 0;
  }

  .footer {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    margin-top: var(--padding);
    gap: var(--padding-large);

    .cancel {
      @include fonts.button_text;
      cursor: pointer;
      color: var(--text-color-discreet);
    }

    .delete {
      @include main.button-dangerous;
    }
  }
}
</style>
