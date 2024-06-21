<script lang="ts" setup>
import { useField, useForm } from 'vee-validate'
import { API_PATH } from '~/types/customFetch'

const { t: $t } = useI18n()
const { fetch } = useCustomFetch()
const toast = useBcToast()
const { handleSubmit, errors } = useForm()
const { value: newEmail } = useField<string>('newEmail', validateEmail)
const { value: confirmEmail } = useField<string>('confirmEmail', validateEmail)

const buttonsDisabled = defineModel<boolean | undefined>({ required: true })

// TODO: Use userValidation.ts
function validateEmail (value: string) : true | string {
  if (!value) {
    return $t('login_and_register.no_email')
  }
  if (!REGEXP_VALID_EMAIL.test(value)) {
    return $t('login_and_register.invalid_email')
  }
  return true
}

const onSubmit = handleSubmit(async (values) => {
  if (!canSubmit.value) {
    return
  }

  buttonsDisabled.value = true
  try {
    await fetch(API_PATH.USER_CHANGE_EMAIL, {
      body: {
        email: values.newEmail
      }
    })
  } catch (error) {
    toast.showError(
      {
        summary: $t('user_settings.email.error.toast_title'),
        group: $t('user_settings.email.error.toast_group'),
        detail: $t('user_settings.email.error.toast_message')
      })
  }
  buttonsDisabled.value = false
})

const newEmailError = ref<string|undefined>(undefined)
const confirmEmailError = ref<string|undefined>(undefined)

const canSubmit = computed(() => !buttonsDisabled.value && newEmail.value && confirmEmail.value && newEmail.value === confirmEmail.value && !Object.keys(errors.value).length)

</script>

<template>
  <form class="email-container" @submit="onSubmit">
    <div class="title">
      {{ $t('user_settings.email.title') }}
    </div>
    <label for="new-email">
      {{ $t('user_settings.email.new') }}
    </label>
    <div class="input-row">
      <InputText
        id="new-email"
        v-model="newEmail"
        :class="{ 'p-invalid': errors?.newEmail }"
        aria-describedby="text-error"
        @focusin="newEmailError = undefined"
        @focusout="newEmailError = errors?.newEmail"
      />
      <div class="p-error">
        {{ newEmailError || '&nbsp;' }}
      </div>
    </div>
    <label for="confirm-email">
      {{ $t('user_settings.email.confirm') }}
    </label>
    <div class="input-row">
      <InputText
        id="confirm-email"
        v-model="confirmEmail"
        :class="{ 'p-invalid': errors?.confirmEmail }"
        aria-describedby="text-error"
        @focusin="confirmEmailError = undefined"
        @focusout="confirmEmailError = errors?.confirmEmail"
      />
      <div class="p-error">
        {{ confirmEmailError || '&nbsp;' }}
      </div>
    </div>
    <div class="button-row">
      <Button type="submit" :disabled="!canSubmit" :label="$t('navigation.save')" />
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
      @include fonts.small_text;
    }
  }

  .button-row {
    display: flex;
    justify-content: flex-end;
  }
}
</style>
