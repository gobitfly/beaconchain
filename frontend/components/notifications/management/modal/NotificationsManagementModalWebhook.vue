<script lang="ts" setup>
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faPaperPlane } from '@fortawesome/pro-solid-svg-icons'
import { useForm } from 'vee-validate'
import { warn } from 'vue'
import { API_PATH } from '~/types/customFetch'

export type WebhookForm = {
  is_discord_webhook_enabled: boolean,
  webhook_url: string,
}
const {
  close, props,
} = useBcDialog<WebhookForm>()

const { t: $t } = useTranslation()

const validationSchema = createSchemaObject({
  is_discord_webhook_enabled: validation.boolean(),
  webhook_url: validation.url($t('validation.url.invalid')),
})

const {
  defineField, errors, handleSubmit, meta, setFieldError, values,
}
  = useForm({
    initialValues: {
      is_discord_webhook_enabled:
        props.value?.is_discord_webhook_enabled || false,
      webhook_url: props.value?.webhook_url || '',
    },
    validationSchema,
  })

const [
  webhook_url,
  webhook_url_attrs,
] = defineField('webhook_url', { validateOnModelUpdate: false })

const [
  is_discord_webhook_enabled,
  is_discord_webhook_enabled_attrs,
]
  = defineField('is_discord_webhook_enabled', { validateOnModelUpdate: false })

const isFormDirty = computed(() => meta.value.dirty)
const isFormValid = computed(() => meta.value.valid)

const toast = useBcToast()
const { fetch } = useCustomFetch()
const handleTestNotification = async () => {
  // 1. could not be implemented as a custom validation rule,
  // as they are always triggerd onMounted (at cast time)
  if (!webhook_url.value && is_discord_webhook_enabled.value) {
    // 1.
    setFieldError('webhook_url', $t('validation.webhook.discord_empty'))
    return
  }
  if (!webhook_url.value && !is_discord_webhook_enabled.value) {
    // 1.
    setFieldError('webhook_url', $t('validation.webhook.empty'))
    return
  }
  if (!isFormValid.value) {
    return
  }
  try {
    if (is_discord_webhook_enabled.value) {
      await fetch(API_PATH.NOTIFICATIONS_TEST_WEBHOOK, {
        body: {
          is_discord_webhook_enabled: is_discord_webhook_enabled.value,
          webhook_url: webhook_url.value,
        },
        method: 'POST',
      })
      toast.showSuccess({ summary: $t('notifications.dashboards.toast.success.test_discord') })
      return
    }
    await fetch(API_PATH.NOTIFICATIONS_TEST_WEBHOOK, {
      body: { webhook_url: webhook_url.value },
      method: 'POST',
    })
    toast.showSuccess({ summary: $t('notifications.dashboards.toast.success.test_webhook_url') })
  }
  catch (error) {
    const summary = is_discord_webhook_enabled.value
      ? $t('notifications.dashboards.toast.error.discord')
      : $t('notifications.dashboards.toast.error.webhook_url')
    toast.showError({ summary })
  }
  warn('Test notification sent', values)
}
const emit = defineEmits<{
  (e: 'save', values: WebhookForm, closeCallback: () => void): void,
}>()

const onSubmit = handleSubmit((values) => {
  if (!isFormDirty.value) {
    close()
    return
  }
  emit('save', values, close)
})

const id = useId()
</script>

<template>
  <h2
    :id
    class="notifications-management-dialog-webhook__header"
  >
    {{ $t("notifications.dashboards.dialog.heading_webhook") }}
  </h2>
  <BaseForm
    v-focustrap
    novalidate
    class="notifications-management-dialog-webhook__form"
    :aria-describedby="id"
    @keydown.esc.stop.prevent="close"
    @submit.prevent="onSubmit"
  >
    <BaseFormRow>
      <BcInputText
        v-model="webhook_url"
        v-bind="webhook_url_attrs"
        :label="$t('notifications.dashboards.dialog.label_webhook_url')"
        :placeholder="$t('notifications.dashboards.dialog.placeholder_webhook')"
        input-width="200px"
        :error="errors.webhook_url"
        type="url"
        should-autoselect
      />
    </BaseFormRow>
    <BaseFormRow>
      <BcInputCheckbox
        v-model="is_discord_webhook_enabled"
        v-bind="is_discord_webhook_enabled_attrs"
        :label="$t('notifications.dashboards.dialog.label_send_via_discord')"
        :error="errors.is_discord_webhook_enabled"
      >
        <template #tooltip>
          <BcTranslation
            keypath="notifications.dashboards.dialog.info_send_via_discord.template"
            linkpath="notifications.dashboards.dialog.info_send_via_discord._link"
            to="https://discord.com/developers/docs/resources/webhook"
          />
        </template>
      </BcInputCheckbox>
    </BaseFormRow>
    <div class="notifications-management-dialog-webhook-footer">
      <BcButton
        font-awesome-icon="faPaperPlane"
        variant="secondary"
        :is-aria-disabled="!isFormValid || !webhook_url"
        @click="handleTestNotification()"
      >
        {{ $t("notifications.dashboards.dialog.button_webhook_test") }}
        <template #icon>
          <FontAwesomeIcon :icon="faPaperPlane" />
        </template>
      </BcButton>
      <BcButton
        class="notifications-management-dialog-webhook__primary-button"
        @click="onSubmit"
      >
        {{ isFormDirty ? $t("navigation.save") : $t("navigation.done") }}
      </BcButton>
    </div>
  </BaseForm>
</template>

<style lang="scss" scoped>
.notifications-management-dialog-webhook__header {
  margin: unset;
  text-align: center;
  font-family: var(--header1_font_family);
  font-size: var(--header1_font_size);
  font-weight: var(--header1_font_weight);
}
.notifications-management-dialog-webhook__form {
  margin-top: var(--padding-large);
}
.notifications-management-dialog-webhook-footer {
  display: flex;
  gap: var(--padding-small);
  justify-content: flex-end;
  margin-top: var(--padding-small);
}
.notifications-management-dialog-webhook__primary-button {
  // otherwise layout will jump when text changes
  min-width: 90px;
}
</style>
