<script setup lang="ts">
import { object as yupObject } from 'yup'
import { useForm } from 'vee-validate'
import { Target } from '~/types/links'
import { tOf } from '~/utils/translation'
import { API_PATH } from '~/types/customFetch'

const { t: $t } = useI18n()
const { fetch } = useCustomFetch()
const toast = useBcToast()

useBcSeo('login_and_register.title_register')

const { handleSubmit, errors, defineField } = useForm({
  validationSchema: yupObject({
    email: emailValidation($t),
    password: passwordValidation($t),
    confirmPassword: confirmPasswordValidation($t, 'password'),
    agreement: checkboxValidation('')
  })
})

const [email, emailAttrs] = defineField('email')
const [password, passwordAttrs] = defineField('password')
const [confirmPassword, confirmPasswordAttrs] = defineField('confirmPassword')
const [agreement, agreementAttrs] = defineField('agreement')

const onSubmit = handleSubmit(async (values) => {
  if (!canSubmit.value) {
    return
  }
  try {
    await fetch(API_PATH.REGISTER, {
      method: 'POST',
      body: {
        email: values.email,
        password: values.password
      }
    })
    await navigateTo('/')
  } catch (error) {
    toast.showError({ summary: $t('login_and_register.error_title'), group: $t('login_and_register.error_register_group'), detail: $t('login_and_register.error_register_message') })
  }
})

const canSubmit = computed(() => email.value && password.value && confirmPassword.value && agreement.value && !Object.keys(errors.value).length)
</script>

<template>
  <BcPageWrapper :minimalist-header="true">
    <div class="page">
      <div class="container">
        <div class="title">
          {{ $t('login_and_register.title_register') }}
        </div>
        <div class="login-invitation">
          {{ $t('login_and_register.already_have_account') }}
          <BcLink to="/login" :target="Target.Internal" class="link">
            {{ $t('login_and_register.login_here') }}
          </BcLink>
        </div>
        <form @submit="onSubmit">
          <div class="input-row">
            <label for="email" class="label">{{ $t('login_and_register.email') }}</label>
            <InputText
              id="email"
              v-model="email"
              v-bind="emailAttrs"
              type="text"
              :placeholder="$t('login_and_register.email')"
              :class="{ 'p-invalid': errors?.email }"
              aria-describedby="text-error"
            />
            <div class="p-error">
              {{ errors?.email }}
            </div>
          </div>
          <div class="input-row">
            <label for="password" class="label">{{ $t('login_and_register.choose_password') }}</label>
            <InputText
              id="password"
              v-model="password"
              v-bind="passwordAttrs"
              type="password"
              :placeholder="$t('login_and_register.password')"
              :class="{ 'p-invalid': errors?.password }"
              aria-describedby="text-error"
            />
            <div class="p-error">
              {{ errors?.password }}
            </div>
          </div>
          <div class="input-row">
            <label for="confirmPassword" class="label">{{ $t('login_and_register.confirm_password') }}</label>
            <InputText
              id="confirmPassword"
              v-model="confirmPassword"
              v-bind="confirmPasswordAttrs"
              type="password"
              :placeholder="$t('login_and_register.password')"
              :class="{ 'p-invalid': errors?.confirmPassword }"
              aria-describedby="text-error"
            />
            <div class="p-error">
              {{ errors?.confirmPassword }}
            </div>
          </div>
          <div class="last-row">
            <div class="input-with-error">
              <div class="agreement">
                <Checkbox
                  v-model="agreement"
                  v-bind="agreementAttrs"
                  input-id="agreement"
                  :binary="true"
                  type="checkbox"
                  class="checkbox"
                  aria-describedby="text-error"
                />
                <div class="text">
                  <label for="agreement">{{ tOf($t, 'login_and_register.please_agree', 0) + ' ' }}</label>
                  <BcLink to="https://storage.googleapis.com/legal.beaconcha.in/tos.pdf" :target="Target.External" class="link">
                    {{ tOf($t, 'login_and_register.please_agree', 1) }}
                  </BcLink>
                  {{ tOf($t, 'login_and_register.please_agree', 2) }}
                  <BcLink to="https://storage.googleapis.com/legal.beaconcha.in/privacy.pdf" :target="Target.External" class="link">
                    {{ tOf($t, 'login_and_register.please_agree', 3) }}
                  </BcLink>
                </div>
              </div>
              <div class="p-error">
                {{ errors?.agreement }}
              </div>
            </div>
            <Button class="button" type="submit" :label="$t('login_and_register.submit_register')" :disabled="!canSubmit" />
          </div>
        </form>
      </div>
    </div>
  </BcPageWrapper>
</template>

<style lang="scss" scoped>
@use "~/assets/css/fonts.scss";

.page {
  display: flex;
  flex-direction: column;

  .container {
    position: relative;
    margin: auto;
    margin-top: 100px;
    @media (max-width: 600px) { // mobile
      margin-top: 0px;
    }
    margin-bottom: 30px;
    padding: var(--padding-large);
    box-sizing: border-box;
    width: min(530px, 100%);

    .title {
      @include fonts.dialog_header;
      margin-bottom: var(--padding-large);
    }

    .login-invitation {
      position: relative;
      @include fonts.small_text;
      margin-bottom: var(--padding-large);
    }

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
        height: 100px;
        .label {
          margin-bottom: 12px;
        }
      }

      .last-row {
        position: relative;
        display: flex;
        margin-top: auto;
        .input-with-error {
          display: flex;
          position: relative;
          margin: auto;
          margin-left: 0;
          .agreement {
            display: flex;
            position: relative;
            flex-direction: row;
            gap: 10px;
            .checkbox {
              margin-top: auto;
              margin-bottom: auto;
            }
            .text {
              @media (max-width: 600px) { // mobile
                font-size: var(--small_text_font_size);
                line-height: 20px;
              }
            }
          }

        }
        .button {
          margin: auto;
          margin-right: 0;
          @media (max-width: 600px) { // mobile
            width: 70px;
            padding: 0;
          }
        }
      }

      .p-error {
        margin-top: var(--padding-small);
        @include fonts.small_text;
        font-weight: var(--roboto-regular);
      }
    }
  }
}
</style>
