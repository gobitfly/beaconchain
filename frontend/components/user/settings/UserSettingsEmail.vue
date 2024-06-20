<script lang="ts" setup>
import { useField, useForm } from 'vee-validate'

const { t: $t } = useI18n()
const toast = useBcToast()
const { handleSubmit, errors } = useForm()
const { value: newEmail } = useField<string>('newEmail', validateEmail)
const { value: confirmEmail } = useField<string>('confirmEmail', validateEmail)

// TODO: This duplicates code from login.vue. Move to a shared file.
function validateEmail (value: string) : true | string {
  if (!value) {
    return $t('login.no_email')
  }
  if (!REGEXP_VALID_EMAIL.test(value)) {
    return $t('login.invalid_email')
  }
  return true
}

const onSubmit = handleSubmit(async (values) => {
  // TODO: remove
  console.log('submitting email form with values:', values)
  await new Promise(resolve => setTimeout(resolve, 1000))

  try {
    // TODO: implement
  } catch (error) {
    toast.showError({ summary: $t('login.error_toast_title'), group: $t('login.error_toast_group'), detail: $t('login.error_toast_message') })
  }
})

const newEmailError = ref<string|undefined>(undefined)
const confirmEmailError = ref<string|undefined>(undefined)

const canSubmit = computed(() => newEmail.value && confirmEmail.value && newEmail.value === confirmEmail.value && !Object.keys(errors.value).length)

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
