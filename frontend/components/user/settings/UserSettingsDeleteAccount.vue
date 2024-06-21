<script lang="ts" setup>
import { BcDialogDelete } from '#components'
import { API_PATH } from '~/types/customFetch'

const dialog = useDialog()
const { t: $t } = useI18n()
const { fetch } = useCustomFetch()

const buttonsDisabled = defineModel<boolean | undefined>({ required: true })

const onDelete = () => {
  dialog.open(BcDialogDelete, {
    data: {
      title: $t('user_settings.delete_account.dialog.title'),
      warning: $t('user_settings.delete_account.dialog.warning'),
      yesLabel: $t('user_settings.delete_account.dialog.yes_label')
    },
    onClose: response => response?.data && deleteAction()
  })
}

const deleteAction = async () => {
  if (buttonsDisabled.value) {
    return
  }

  buttonsDisabled.value = true
  await fetch(API_PATH.USER_DELETE)
  await navigateTo('/')
}

</script>

<template>
  <div class="subscriptions-container">
    <div class="title">
      {{ $t('user_settings.delete_account.title') }}
    </div>
    <div class="warning-row">
      <div class="warning">
        {{ $t('user_settings.delete_account.warning') }}
      </div>
      <Button :label="$t('user_settings.delete_account.button')" :disabled="buttonsDisabled" class="delete-button" @click="onDelete" />
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use '~/assets/css/main.scss';
@use '~/assets/css/fonts.scss';

.subscriptions-container {
  display: flex;
  flex-direction: column;
  gap: var(--padding);

  padding: var(--padding-large);
  @include main.container;

  .title {
    @include fonts.dialog_header;
    margin-bottom: 9px;
  }

  .warning-row {
    display: flex;
    justify-content: space-between;
    align-items: center;

    .warning {
      @include fonts.subtitle_text;
    }

    .delete-button {
      @include main.button-dangerous;
    }
  }
}
</style>
