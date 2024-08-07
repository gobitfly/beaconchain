<script lang="ts" setup>
import { object as yupObject } from 'yup'
import { useForm } from 'vee-validate'
import { API_PATH } from '~/types/customFetch'

const { t: $t } = useTranslation()
const { fetch } = useCustomFetch()
const toast = useBcToast()

const { handleSubmit, errors, defineField } = useForm({
  validationSchema: yupObject({
    password: passwordValidation($t),
    newEmail: emailValidation($t),
    confirmEmail: confirmEmailValidation($t, 'newEmail'),
  }),
})

const [password, passwordAttrs] = defineField('password')
const [newEmail, newEmailAttrs] = defineField('newEmail')
const [confirmEmail, confirmEmailAttrs] = defineField('confirmEmail')

const buttonsDisabled = defineModel<boolean | undefined>({ required: true })

const onSubmit = handleSubmit(async (values, { resetForm }) => {
  if (!canSubmit.value) {
    return
  }

  buttonsDisabled.value = true
  try {
    await fetch(API_PATH.USER_CHANGE_EMAIL, {
      body: {
        password: values.password,
        email: values.newEmail,
      },
    })
    toast.showSuccess(
      {
        summary: $t('user_settings.email.success.toast_title'),
        group: $t('user_settings.email.success.toast_group'),
        detail: $t('user_settings.email.success.toast_message'),
      })
    resetForm()
  }
  catch (error) {
    toast.showError(
      {
        summary: $t('user_settings.email.error.toast_title'),
        group: $t('user_settings.email.error.toast_group'),
        detail: $t('user_settings.email.error.toast_message'),
      })
  }
  buttonsDisabled.value = false
})

const canSubmit = computed(() => !buttonsDisabled.value && newEmail.value && confirmEmail.value && newEmail.value === confirmEmail.value && password.value && !Object.keys(errors.value).length)
</script>

<template>
  <form
    class="email-container"
    @submit="onSubmit"
  >
    <div class="title">
      {{ $t('user_settings.email.title') }}
    </div>
    <label for="password">
      {{ $t('user_settings.email.password') }}
    </label>
    <div class="input-row">
      <InputText
        id="password"
        v-model="password"
        v-bind="passwordAttrs"
        type="password"
        :class="{ 'p-invalid': errors?.password }"
        aria-describedby="text-error"
      />
      <div class="p-error">
        {{ errors?.password }}
      </div>
    </div>
    <label for="new-email">
      {{ $t('user_settings.email.new') }}
    </label>
    <div class="input-row">
      <InputText
        id="new-email"
        v-model="newEmail"
        v-bind="newEmailAttrs"
        type="text"
        :class="{ 'p-invalid': errors?.newEmail }"
        aria-describedby="text-error"
      />
      <div class="p-error">
        {{ errors?.newEmail }}
      </div>
    </div>
    <label for="confirm-email">
      {{ $t('user_settings.email.confirm') }}
    </label>
    <div class="input-row">
      <InputText
        id="confirm-email"
        v-model="confirmEmail"
        v-bind="confirmEmailAttrs"
        type="text"
        :class="{ 'p-invalid': errors?.confirmEmail }"
        aria-describedby="text-error"
      />
      <div class="p-error">
        {{ errors?.confirmEmail }}
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
@use '~/assets/css/main.scss';
@use '~/assets/css/fonts.scss';

.email-container {
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
