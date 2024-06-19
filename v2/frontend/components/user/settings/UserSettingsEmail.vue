<script lang="ts" setup>
import { useField } from 'vee-validate'

const { t: $t } = useI18n()

const { value: newEmail } = useField<string>('newEmail', validateEmail)
const { value: confirmEmail } = useField<string>('confirmEmail', validateEmail)
const saveDisabled = computed(() => {
  return newEmail.value === '' || confirmEmail.value === '' || newEmail.value !== confirmEmail.value
})

// TODO: implement properly (also use vor saveDisabled)
function validateEmail (value: string) : true | string {
  if (!value) {
    return $t('login.no_email')
  }
  if (!REGEXP_VALID_EMAIL.test(value)) {
    return $t('login.invalid_email')
  }
  return true
}

</script>

<template>
  <div class="email-container">
    <div class="title">
      {{ $t('user_settings.email.title') }}
    </div>
    <label for="new-email">
      {{ $t('user_settings.email.new') }}
    </label>
    <InputText id="new-email" v-model="newEmail" />
    <label for="confirm-email">
      {{ $t('user_settings.email.confirm') }}
    </label>
    <InputText id="confirm-email" v-model="confirmEmail" />
    <div class="button-row">
      <Button :disabled="saveDisabled" :label="$t('navigation.save')" />
    </div>
  </div>
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

  input {
    margin-bottom: 9px;
  }

  .button-row {
    display: flex;
    justify-content: flex-end;
  }
}
</style>
