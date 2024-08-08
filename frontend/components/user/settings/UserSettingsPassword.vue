<script lang="ts" setup>
import { object as yupObject } from 'yup'
import { useForm } from 'vee-validate'
import { API_PATH } from '~/types/customFetch'

const { t: $t } = useTranslation()
const { fetch } = useCustomFetch()
const toast = useBcToast()

const { defineField, errors, handleSubmit } = useForm({
  validationSchema: yupObject({
    confirmPassword: confirmPasswordValidation($t, 'newPassword'),
    newPassword: newPasswordValidation($t, 'oldPassword'),
    oldPassword: passwordValidation($t),
  }),
})

const [oldPassword, oldPasswordAttrs] = defineField('oldPassword')
const [newPassword, newPasswordAttrs] = defineField('newPassword')
const [confirmPassword, confirmPasswordAttrs] = defineField('confirmPassword')

const buttonsDisabled = defineModel<boolean | undefined>({ required: true })

const onSubmit = handleSubmit(async (values, { resetForm }) => {
  if (!canSubmit.value) {
    return
  }

  buttonsDisabled.value = true
  try {
    await fetch(API_PATH.USER_CHANGE_PASSWORD, {
      body: {
        password: values.newPassword,
      },
    })
    toast.showSuccess({
      detail: $t('user_settings.password.success.toast_message'),
      group: $t('user_settings.password.success.toast_group'),
      summary: $t('user_settings.password.success.toast_title'),
    })
    resetForm()
  }
  catch (error) {
    toast.showError({
      detail: $t('user_settings.password.error.toast_message'),
      group: $t('user_settings.password.error.toast_group'),
      summary: $t('user_settings.password.error.toast_title'),
    })
  }
  buttonsDisabled.value = false
})

const canSubmit = computed(
  () =>
    !buttonsDisabled.value
    && oldPassword.value
    && newPassword.value
    && confirmPassword.value
    && newPassword.value === confirmPassword.value
    && !Object.keys(errors.value).length,
)
</script>

<template>
  <form
    class="password-container"
    @submit="onSubmit"
  >
    <div class="title">
      {{ $t("user_settings.password.title") }}
    </div>
    <label for="old-password">
      {{ $t("user_settings.password.old") }}
    </label>
    <div class="input-row">
      <InputText
        id="old-password"
        v-model="oldPassword"
        v-bind="oldPasswordAttrs"
        type="password"
        :class="{ 'p-invalid': errors?.oldPassword }"
        aria-describedby="text-error"
      />
      <div class="p-error">
        {{ errors?.oldPassword }}
      </div>
    </div>
    <label for="new-password">
      {{ $t("user_settings.password.new") }}
    </label>
    <div class="input-row">
      <InputText
        id="new-password"
        v-model="newPassword"
        v-bind="newPasswordAttrs"
        type="password"
        :class="{ 'p-invalid': errors?.newPassword }"
        aria-describedby="text-error"
      />
      <div class="p-error">
        {{ errors?.newPassword }}
      </div>
    </div>
    <label for="confirm-password">
      {{ $t("user_settings.password.confirm") }}
    </label>
    <div class="input-row">
      <InputText
        id="confirm-password"
        v-model="confirmPassword"
        v-bind="confirmPasswordAttrs"
        type="password"
        :class="{ 'p-invalid': errors?.confirmPassword }"
        aria-describedby="text-error"
      />
      <div class="p-error">
        {{ errors?.confirmPassword }}
      </div>
    </div>
    <div class="button-row">
      <Button
        type="submit"
        :disabled="!canSubmit"
        :label="$t('navigation.save')"
      />
    </div>
  </form>
</template>

<style lang="scss" scoped>
@use "~/assets/css/main.scss";
@use "~/assets/css/fonts.scss";

.password-container {
  display: flex;
  flex-direction: column;
  gap: var(--padding-small);

  @include main.container;
  padding: var(--padding-large);

  .title {
    @include fonts.dialog_header;
    margin-bottom: 15px;
  }

  label {
    @include fonts.small_text;
  }

  .input-row {
    input {
      width: 100%;
    }

    .p-error {
      min-height: 17px;
      @include fonts.small_text;
    }
  }

  .button-row {
    display: flex;
    justify-content: flex-end;
  }
}
</style>
