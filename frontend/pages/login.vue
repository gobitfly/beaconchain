<script setup lang="ts">
import { object as yupObject } from 'yup'
import { useForm } from 'vee-validate'
import { useUserStore } from '~/stores/useUserStore'
import { Target } from '~/types/links'
import { provideMobileAuthParams, handleMobileAuth } from '~/utils/mobileAuth'

const { t: $t } = useTranslation()
const { doLogin } = useUserStore()
const toast = useBcToast()
const route = useRoute()

useBcSeo('login_and_register.title_login')

const { handleSubmit, errors, defineField } = useForm({
  validationSchema: yupObject({
    email: emailValidation($t),
    password: passwordValidation($t),
  }),
})

const [email, emailAttrs] = defineField('email')
const [password, passwordAttrs] = defineField('password')

const onSubmit = handleSubmit(async (values) => {
  try {
    await doLogin(values.email, values.password)

    if (handleMobileAuth(route.query)) {
      return
    }

    await navigateTo('/')
  }
  catch (error) {
    password.value = ''
    toast.showError({
      summary: $t('login_and_register.error_title'),
      group: $t('login_and_register.error_login_group'),
      detail: $t('login_and_register.error_login_message'),
    })
  }
})

const canSubmit = computed(() => email.value && password.value && !Object.keys(errors.value).length)

const registerLink = computed(() => {
  return provideMobileAuthParams(route.query, '/register')
})
</script>

<template>
  <BcPageWrapper :minimalist-header="true">
    <div class="page">
      <div class="container">
        <div class="title">
          {{ $t("login_and_register.title_login") }}
        </div>
        <form @submit="onSubmit">
          <div class="input-row">
            <label
              for="email"
              class="label"
            >{{
              $t("login_and_register.email")
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
            >
              <div>{{ $t("login_and_register.password") }}</div>
              <div class="right-cell">
                <BcLink
                  to="/requestReset"
                  :target="Target.Internal"
                  class="link"
                >
                  {{ $t("login_and_register.forgotten") }}
                </BcLink>
              </div>
            </label>
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
              {{ $t("login_and_register.dont_have_account") }}
              <BcLink
                :to="registerLink"
                :target="Target.Internal"
                class="link"
              >
                {{ $t("login_and_register.signup_here") }}
              </BcLink>
            </div>
            <Button
              class="button"
              type="submit"
              :label="$t('login_and_register.submit_login')"
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
          display: flex;
          flex-direction: row;
          margin-bottom: 12px;
          .right-cell {
            margin-left: auto;
            @include fonts.small_text;
          }
        }
      }

      .last-row {
        position: relative;
        display: flex;
        margin-top: auto;
        .account-invitation {
          @include fonts.small_text;
          position: relative;
          margin: auto;
          margin-left: 0;
        }
        .button {
          margin-left: auto;
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
