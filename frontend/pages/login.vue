<script setup lang="ts">
import { object as yupObject } from 'yup'
import { useForm } from 'vee-validate'
import { useUserStore } from '~/stores/useUserStore'
import { Target } from '~/types/links'

const { t: $t } = useI18n()
const { doLogin } = useUserStore()
const toast = useBcToast()

useBcSeo('login_and_register.title_login')

const { handleSubmit, errors, defineField } = useForm({
  validationSchema: yupObject({
    email: emailValidation($t),
    password: passwordValidation($t)
  })
})

const [email, emailAttrs] = defineField('email')
const [password, passwordAttrs] = defineField('password')

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
            v-bind="emailAttrs"
            type="text"
            :class="{ 'p-invalid': errors?.email }"
            aria-describedby="text-error"
          />
          <div class="p-error">
            {{ errors?.email }}
          </div>
        </div>
        <div class="input-row">
          <label for="password" class="label">{{ $t('login_and_register.password') }}</label>
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
      min-height: 17px;
      @include fonts.small_text;
    }
  }
}
</style>
