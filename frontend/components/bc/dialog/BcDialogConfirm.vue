<script setup lang="ts">

interface Props {
  title?: string
  question?: string
  noLabel?: string // defaults to "No"
  yesLabel?: string // defaults to "Yes"
  severity?: 'default' | 'danger'
}
const { props, dialogRef } = useBcDialog<Props>({ showHeader: false })
const { t: $t } = useTranslation()

const noLabel = computed(() => props.value?.noLabel || $t('navigation.no'))
const yesLabel = computed(() => props.value?.yesLabel || $t('navigation.yes'))

const closeDialog = (response: boolean) => {
  dialogRef?.value.close(response)
}
</script>

<template>
  <div class="content">
    <div v-if="props?.title" class="title">
      {{ props?.title }}
    </div>
    <div v-if="props?.question" class="question">
      {{ props?.question }}
    </div>
    <div class="footer">
      <Button v-if="props?.severity !== 'danger'" type="button" :label="noLabel" @click="closeDialog(false)" />
      <div v-else class="discreet-button" @click="closeDialog(false)">
        {{ noLabel }}
      </div>
      <Button type="button" :severity="props?.severity === 'danger' ? `danger` : undefined" :label="yesLabel" @click="closeDialog(true)" />
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
    @include fonts.small_text;
    font-weight: var(--roboto-medium);
    margin: var(--padding) 0;
  }

  .footer {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    margin-top: var(--padding);
    gap: var(--padding);

    .discreet-button{
      @include fonts.button_text;
      cursor: pointer;
      color: var(--text-color-discreet);
      margin-right: var(--padding);
    }
  }
}
</style>
