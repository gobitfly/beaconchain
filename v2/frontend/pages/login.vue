<script setup lang="ts">
import { useField, useForm } from 'vee-validate'
import { useUserStore } from '~/stores/useUserStore'
import { REGEXP_VALID_EMAIL } from '~/utils/regexp'
import { Target } from '~/types/links'

const { t } = useI18n()
const { doLogin } = useUserStore()
const toast = useBcToast()

const { handleSubmit, errors } = useForm()
const { value: email } = useField<string>('email', validateAddress)
const { value: password } = useField<string>('password', validatePassword)

function validateAddress (value: string) : true|string {
  if (!value) {
    return t('login.no_email')
  }
  if (!REGEXP_VALID_EMAIL.test(value)) {
    return t('login.invalid_email')
  }
  return true
}

function validatePassword (value: string) : true|string {
  if (!value) {
    return t('login.no_password')
  }
  if (value.length < 5) {
    return t('login.invalid_password')
  }
  return true
}

const onSubmit = handleSubmit(async (values) => {
  try {
    await doLogin(values.email, values.password)
    await navigateTo('/')
  } catch (error) {
    password.value = ''
    toast.showError({ summary: t('login.error_toast_title'), group: t('login.error_toast_group'), detail: t('login.error_toast_message') })
  }
})

const canSubmit = computed(() => email.value && password.value && !Object.keys(errors.value).length)
const addressError = ref<string|undefined>(undefined)
const passwordError = ref<string|undefined>(undefined)
</script>

<template>
  <BcPageWrapper>
    <div class="container">
      <form @submit="onSubmit">
        <div class="input-row">
          <label for="email" class="label">{{ $t('login.email') }}</label>
          <InputText
            id="email"
            v-model="email"
            type="text"
            :class="{ 'p-invalid': errors?.email }"
            aria-describedby="text-error"
            @focus="addressError = undefined"
            @blur="addressError = errors?.email"
          />
          <div class="p-error">
            {{ addressError || '&nbsp;' }}
          </div>
        </div>
        <div class="input-row">
          <label for="password" class="label">{{ $t('login.password') }}</label>
          <InputText
            id="password"
            v-model="password"
            type="password"
            :class="{ 'p-invalid': errors?.password }"
            aria-describedby="text-error"
            @focus="passwordError = undefined"
            @blur="passwordError = errors?.password"
          />
          <div class="p-error">
            {{ passwordError || '&nbsp;' }}
          </div>
        </div>
        <div class="last-row">
          <div class="account-invitation">
            {{ t('login.dont_have_account') }}<br>
            <NuxtLink to="/register" :target="Target.Internal" class="link">
              {{ t('login.signup_here') }}
            </NuxtLink>
          </div>
          <Button type="submit" :label="$t('login.submit')" :disabled="!canSubmit" />
        </div>
      </form>
    </div>
  </BcPageWrapper>
</template>

<style lang="scss" scoped>
@use "~/assets/css/fonts.scss";

.container {
  position: relative;
  width: 300px;
  height: 230px;
  margin: auto;
  margin-top: 60px;
  margin-bottom: 50px;
  padding: var(--padding);

  form {
    @include fonts.standard_text;
    position: relative;
    display: flex;
    flex-direction: column;
    width: 100%;
    height: 100%;

    .input-row {
      position: relative;
      display: flex;
      flex-direction: column;
      margin: var(--padding) auto 0px auto;
      width: 80%;

      &:first-child {
        padding-top: var(--padding-small)
      }
      .label {
        margin-bottom: var(--padding-small);
      }
    }

    .last-row {
      position: relative;
      display: flex;
      margin-top: auto;

      .account-invitation {
        position: relative;
        @include fonts.small_text;
        margin-right: auto;
      }
    }

    .p-error {
      @include fonts.small_text;
    }
  }
}
</style>
