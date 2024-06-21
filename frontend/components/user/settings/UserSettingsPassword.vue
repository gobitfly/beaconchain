<script lang="ts" setup>
import { useField, useForm } from 'vee-validate'
import { API_PATH } from '~/types/customFetch'

const { t: $t } = useI18n()
const { fetch } = useCustomFetch()
const toast = useBcToast()
const { handleSubmit, errors } = useForm()
const { value: oldPassword } = useField<string>('oldPassword', validatePassword)
const { value: newPassword } = useField<string>('newPassword', validatePassword) // TODO: This should also validate that new != old (wait for shared file, see below)
const { value: confirmPassword } = useField<string>('confirmPassword', validatePassword)

const buttonsDisabled = defineModel<boolean | undefined>({ required: true })

// TODO: This duplicates code from login.vue. Move to a shared file.
// Shared file will be created in a different PR. Wait until it can be merged.
function validatePassword (value: string) : true | string {
  if (!value) {
    return $t('login.no_password')
  }
  if (value.length < 5) {
    return $t('login.invalid_password')
  }
  return true
}

const onSubmit = handleSubmit(async (values) => {
  if (!canSubmit.value) {
    return
  }

  buttonsDisabled.value = true
  try {
    await fetch(API_PATH.USER_CHANGE_PASSWORD, {
      body: {
        password: values.newPassword
      }
    })
  } catch (error) {
    toast.showError(
      {
        summary: $t('user_settings.password.error.toast_title'),
        group: $t('user_settings.password.error.toast_group'),
        detail: $t('user_settings.password.error.toast_message')
      })
  }
  buttonsDisabled.value = false
})

const oldPasswordError = ref<string|undefined>(undefined)
const newPasswordError = ref<string|undefined>(undefined)
const confirmPasswordError = ref<string|undefined>(undefined)

const canSubmit = computed(() => !buttonsDisabled.value && oldPassword.value && newPassword.value && confirmPassword.value && newPassword.value === confirmPassword.value && !Object.keys(errors.value).length)

</script>

<template>
  <form class="password-container" @submit="onSubmit">
    <div class="title">
      {{ $t('user_settings.password.title') }}
    </div>
    <label for="old-password">
      {{ $t('user_settings.password.old') }}
    </label>
    <div class="input-row">
      <InputText
        id="old-password"
        v-model="oldPassword"
        type="password"
        :class="{ 'p-invalid': errors?.oldPassword }"
        aria-describedby="text-error"
        @focusin="oldPasswordError = undefined"
        @focusout="oldPasswordError = errors?.oldPassword"
      />
      <div class="p-error">
        {{ oldPasswordError || '&nbsp;' }}
      </div>
    </div>
    <label for="new-password">
      {{ $t('user_settings.password.new') }}
    </label>
    <div class="input-row">
      <InputText
        id="new-password"
        v-model="newPassword"
        type="password"
        :class="{ 'p-invalid': errors?.newPassword }"
        aria-describedby="text-error"
        @focusin="newPasswordError = undefined"
        @focusout="newPasswordError = errors?.newPassword"
      />
      <div class="p-error">
        {{ newPasswordError || '&nbsp;' }}
      </div>
    </div>
    <label for="confirm-password">
      {{ $t('user_settings.password.confirm') }}
    </label>
    <div class="input-row">
      <InputText
        id="confirm-password"
        v-model="confirmPassword"
        type="password"
        :class="{ 'p-invalid': errors?.confirmPassword }"
        aria-describedby="text-error"
        @focusin="confirmPasswordError = undefined"
        @focusout="confirmPasswordError = errors?.confirmPassword"
      />
      <div class="p-error">
        {{ confirmPasswordError || '&nbsp;' }}
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
      @include fonts.small_text;
    }
  }

  .button-row {
    display: flex;
    justify-content: flex-end;
  }
}
</style>
