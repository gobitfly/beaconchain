<script setup lang="ts">
import { useField, useForm } from 'vee-validate'
import { useUserStore } from '~/stores/useUserStore'
import { Target } from '~/types/links'
import { setTranslator, validateEmailAddress, validatePassword } from '~/utils/userValidation'

const { t: $t } = useI18n()
const { doLogin } = useUserStore()
const toast = useBcToast()

useBcSeo('login_and_register.title_login')

const { handleSubmit, errors } = useForm()
const { value: email } = useField<string>('email', value => validateEmailAddress(value))
const { value: password } = useField<string>('password', value => validatePassword(value))

setTranslator($t)

const onSubmit = handleSubmit(async (values) => {
  try {
    await doLogin(values.email, values.password)
    await navigateTo('/')
  } catch (error) {
    password.value = ''
    toast.showError({ summary: $t('login_and_register.error_title'), group: $t('login_and_register.error_login_group'), detail: $t('login_and_register.error_login_message') })
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
          <label for="email" class="label">{{ $t('login_and_register.email') }}</label>
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
          <label for="password" class="label">{{ $t('login_and_register.password') }}</label>
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
            {{ $t('login_and_register.dont_have_account') }}<br>
            <BcLink to="/register" :target="Target.Internal" class="link">
              {{ $t('login_and_register.signup_here') }}
            </BcLink>
          </div>
          <Button class="button" type="submit" :label="$t('login_and_register.submit_login')" :disabled="!canSubmit" />
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
  height: 240px;
  margin: auto;
  margin-top: 100px;
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
        padding-top: 5px;
      }
      .label {
        margin-bottom: 8px;
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
      .button {
        margin-top: auto;
      }
    }

    .p-error {
      @include fonts.small_text;
    }
  }
}
</style>
