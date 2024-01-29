<script setup lang="ts">
import { useToast } from 'primevue/usetoast'
import { useField, useForm } from 'vee-validate'
import { useUserStore } from '~/stores/useUserStore'

const { doLogin } = useUserStore()

const { handleSubmit, resetForm, errors, values } = useForm()
const { value: email } = useField('email', validateField)
const { value: password } = useField('password', validateField)
const toast = useToast()

function validateField (value?: string) {
  if (!value) {
    return 'Input required.'
  }

  return true
}

const hasErrors = computed(() => {
  return !!errors.value && !!Object.values(errors.value).filter(val => !!val).length
})

const inputValid = computed(() => {
  return !!values.email?.length && !!values.password?.length && !hasErrors.value
})

const onSubmit = handleSubmit(async (values) => {
  if (inputValid.value) {
    await doLogin(values.email, values.password)

    toast.add({ severity: 'info', summary: 'Form Submitted', detail: `user: ${values.email} pw: ${values.password}`, life: 3000 })
    resetForm()
  }
})

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
          <Button type="submit" :label="$t('login.submit')" :disabled="!inputValid" />
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
