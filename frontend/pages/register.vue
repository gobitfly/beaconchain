<script setup lang="ts">
import { object as yupObject } from 'yup'
import { useForm } from 'vee-validate'
import { Target } from '~/types/links'
import { tOf } from '~/utils/translation'
import { API_PATH } from '~/types/customFetch'
import {
  handleMobileAuth, provideMobileAuthParams,
} from '~/utils/mobileAuth'

const { t: $t } = useTranslation()
const { fetch } = useCustomFetch()
const toast = useBcToast()
const route = useRoute()
const { promoCode } = usePromoCode()

useBcSeo('login_and_register.title_register')

const {
  defineField, errors, handleSubmit,
} = useForm({
  validationSchema: yupObject({
    agreement: checkboxValidation(''),
    confirmPassword: confirmPasswordValidation($t, 'password'),
    email: emailValidation($t),
    password: passwordValidation($t),
  }),
})

const [
  email,
  emailAttrs,
] = defineField('email')
const [
  password,
  passwordAttrs,
] = defineField('password')
const [
  confirmPassword,
  confirmPasswordAttrs,
] = defineField('confirmPassword')
const [
  agreement,
  agreementAttrs,
] = defineField('agreement')

const onSubmit = handleSubmit(async (values) => {
  if (!canSubmit.value) {
    return
  }
  try {
    await fetch(API_PATH.REGISTER, {
      body: {
        email: values.email,
        password: values.password,
      },
      method: 'POST',
    })
    if (handleMobileAuth(route.query)) {
      return
    }
    if (promoCode) {
      await navigateTo({
        path: '/pricing', query: { promoCode },
      })
    }
    else {
      await navigateTo('/')
    }
  }
  catch (error) {
    toast.showError({
      detail: $t('auth.login_and_register.error_register_message'),
      group: $t('auth.login_and_register.error_register_group'),
      summary: $t('auth.login_and_register.error_title'),
    })
  }
})

const canSubmit = computed(
  () =>
    email.value
    && password.value
    && confirmPassword.value
    && agreement.value
    && !Object.keys(errors.value).length,
)

const loginLink = computed(() => {
  return provideMobileAuthParams(route.query, '/login')
})
</script>

<template>
  <BcPageWrapper :minimalist-header="true">
    <div class="page">
      <div class="container">
        <div class="title">
          {{ $t("auth.login_and_register.title_register") }}
        </div>
        <div class="login-invitation">
          {{ $t("auth.login_and_register.already_have_account") }}
          <BcLink
            :to="loginLink"
            :target="Target.Internal"
            class="link"
          >
            {{ $t("auth.login_and_register.login_here") }}
          </BcLink>
        </div>
        <form @submit="onSubmit">
          <div class="input-row">
            <label
              for="email"
              class="label"
            >{{
              $t("auth.login_and_register.email")
            }}</label>
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
            <label
              for="password"
              class="label"
            >{{
              $t("auth.login_and_register.choose_password")
            }}</label>
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
          <div class="input-row">
            <label
              for="confirmPassword"
              class="label"
            >{{
              $t("auth.login_and_register.confirm_password")
            }}</label>
            <InputText
              id="confirmPassword"
              v-model="confirmPassword"
              v-bind="confirmPasswordAttrs"
              type="password"
              :class="{ 'p-invalid': errors?.confirmPassword }"
              aria-describedby="text-error"
            />
            <div class="p-error">
              {{ errors?.confirmPassword }}
            </div>
          </div>
          <div class="last-row">
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
                <label for="agreement">{{
                  tOf($t, "auth.login_and_register.please_agree", 0) + " "
                }}</label>
                <BcLink
                  to="https://storage.googleapis.com/legal.beaconcha.in/tos.pdf"
                  :target="Target.External"
                  class="link"
                >
                  {{ tOf($t, "auth.login_and_register.please_agree", 1) }}
                </BcLink>
                {{ tOf($t, "auth.login_and_register.please_agree", 2) }}
                <BcLink
                  to="https://storage.googleapis.com/legal.beaconcha.in/privacy.pdf"
                  :target="Target.External"
                  class="link"
                >
                  {{ tOf($t, "auth.login_and_register.please_agree", 3) }}
                </BcLink>
              </div>
            </div>
            <Button
              class="button"
              type="submit"
              :label="$t('auth.login_and_register.submit_register')"
              :disabled="!canSubmit"
            />
          </div>
        </form>
      </div>
    </div>
  </BcPageWrapper>
</template>

<style lang="scss" scoped>
@use "~/assets/css/fonts.scss";

.page {
  .container {
    position: relative;
    margin: auto;
    margin-top: 100px;
    margin-bottom: 30px;
    padding: var(--padding-large);
    box-sizing: border-box;
    width: min(530px, 100%);
    @media (max-width: 600px) {
      // mobile
      margin-top: 0px;
    }

    .title {
      @include fonts.dialog_header;
      margin-bottom: var(--padding-large);
    }

    .login-invitation {
      @include fonts.small_text;
      position: relative;
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
        flex-direction: row;
        margin-top: auto;
        .agreement {
          display: flex;
          position: relative;
          margin: auto;
          margin-left: 0;
          gap: 10px;
          .checkbox {
            margin-top: auto;
            margin-bottom: auto;
          }
          .text {
            @media (max-width: 600px) {
              // mobile
              font-size: var(--small_text_font_size);
              line-height: 20px;
            }
          }
        }
        .button {
          margin: auto;
          margin-right: 0;
          @media (max-width: 600px) {
            // mobile
            width: 70px;
            padding: 0;
          }
        }
      }

      .p-error {
        @include fonts.small_text;
        margin-top: var(--padding-small);
        font-weight: var(--roboto-regular);
      }
    }
  }
}
</style>
