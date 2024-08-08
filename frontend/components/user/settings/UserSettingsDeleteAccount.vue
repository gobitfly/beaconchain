<script lang="ts" setup>
import { BcDialogConfirm } from '#components'
import { API_PATH } from '~/types/customFetch'

const dialog = useDialog()
const { t: $t } = useTranslation()
const { fetch } = useCustomFetch()
const { user } = useUserStore()

const buttonsDisabled = defineModel<boolean | undefined>({ required: true })

const onDelete = () => {
  dialog.open(BcDialogConfirm, {
    data: {
      noLabel: $t('user_settings.delete_account.dialog.no_label'),
      question: $t('user_settings.delete_account.dialog.warning', {
        email: user.value?.email || $t('common.unavailable'),
      }),
      severity: 'danger',
      title: $t('user_settings.delete_account.dialog.title'),
      yesLabel: $t('user_settings.delete_account.dialog.yes_label'),
    },
    onClose: response => response?.data && deleteAction(),
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
      {{ $t("user_settings.delete_account.title") }}
    </div>
    <div class="content-container">
      <div class="warning">
        {{ $t("user_settings.delete_account.warning") }}
      </div>
      <div class="button-container">
        <Button
          :label="$t('user_settings.delete_account.button')"
          :disabled="buttonsDisabled"
          severity="danger"
          class="delete-button"
          @click="onDelete"
        />
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/main.scss";
@use "~/assets/css/fonts.scss";

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

  .content-container {
    display: flex;
    justify-content: space-between;
    align-items: center;

    .warning {
      @include fonts.subtitle_text;
    }

    .button-container {
      .delete-button {
        flex-shrink: 0;
      }
    }

    @media (max-width: 730px) {
      flex-direction: column;
      align-items: flex-start;
      gap: var(--padding-large);

      .button-container {
        width: 100%;
        display: flex;
        justify-content: flex-end;
      }
    }
  }
}
</style>
