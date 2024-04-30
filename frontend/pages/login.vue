<script setup lang="ts">
import { useField, useForm } from 'vee-validate'
import { string as yupString } from 'yup'
import { useUserStore } from '~/stores/useUserStore'

const { t } = useI18n()
const { doLogin } = useUserStore()

const { handleSubmit, errors } = useForm()
const { value: email } = useField<string>('email', yupString().email(t('login.invalid_email')).required(t('login.no_email')))
const { value: password } = useField<string>('password', validatePassword)

function validatePassword (value: string) : true|string {
  return !!value || t('login.no_password')
}

const onSubmit = handleSubmit(async (values) => {
  await doLogin(values.email, values.password)
  await navigateTo('/')
})

const canSubmit = computed(() => !!email.value && !!password.value && !Object.keys(errors.value).length)
</script>

<template>
  <BcPageWrapper>
    <div class="container">
      <form @submit="onSubmit">
        <div class="input_row">
          <label for="email">{{ $t('login.email') }}</label>
          <InputText
            id="email"
            v-model="email"
            type="text"
            :class="{ 'p-invalid': errors?.email }"
            aria-describedby="text-error"
          />
          <small id="text-error" class="p-error">{{ errors?.email || '&nbsp;' }}</small>
        </div>
        <div class="input_row">
          <label for="password">{{ $t('login.password') }}</label>
          <InputText
            id="password"
            v-model="password"
            type="password"
            :class="{ 'p-invalid': errors?.password }"
            aria-describedby="text-error"
          />
          <small id="text-error" class="p-error">{{ errors?.password || '&nbsp;' }}</small>
        </div>
        <div class="botton_row">
          <Button type="submit" :label="$t('login.submit')" :disabled="!canSubmit" />
        </div>
      </form>
      <Toast />
    </div>
  </BcPageWrapper>
</template>

<style lang="scss" scoped>
.container {
  display: flex;
  justify-content: center;
  align-content: center;

  form {
    max-width: 50%;

    .input_row {
      display: flex;
      flex-direction: column;
    }

    .botton_row {
      display: flex;
      justify-content: flex-end;
    }
  }
}
</style>
