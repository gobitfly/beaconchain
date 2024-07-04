<script setup lang="ts">
import { object as yupObject } from 'yup'
import { useField, useForm } from 'vee-validate'
import { Target } from '~/types/links'
import { tOf } from '~/utils/translation'
import { API_PATH } from '~/types/customFetch'
import { setTranslator, validateAgreement } from '~/utils/userValidation'

const { t: $t } = useI18n()
const { fetch } = useCustomFetch()
const toast = useBcToast()

useBcSeo('login_and_register.title_register')

const { value: agreement } = useField<boolean>('agreement', validateAgreement)

setTranslator($t)

const { handleSubmit, errors, defineField } = useForm({
  validationSchema: yupObject({
    email: emailValidation($t),
    password: passwordValidation($t),
    confirmPassword: confirmPasswordValidation($t, 'password')
  })
})

const [email, emailAttrs] = defineField('email')
const [password, passwordAttrs] = defineField('password')
const [confirmPassword, confirmPasswordAttrs] = defineField('confirmPassword')

const onSubmit = handleSubmit(async (values) => {
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
const agreementError = ref<string|undefined>(undefined)
</script>

<template>
  <BcPageWrapper>
    <div class="content">
      <div class="caption">
        <div class="title">
          {{ $t('login_and_register.text1_register') }}
        </div>
        <div class="purpose">
          {{ $t('login_and_register.text2_register') }}
        </div>
      </div>
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
            <label for="password" class="label">{{ $t('login_and_register.choose_password') }}</label>
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
            <label for="confirmPassword" class="label">{{ $t('login_and_register.confirm_password') }}</label>
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
          <div class="input-row">
            <div class="agreement">
              <Checkbox
                v-model="agreement"
                input-id="agreement"
                :binary="true"
                class="checkbox"
                @focus="agreementError = undefined"
                @blur="agreementError = errors?.agreement"
              />
              <div>
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
              {{ agreementError || '&nbsp;' }}
            </div>
          </div>
          <div class="last-row">
            <div class="login-invitation">
              {{ $t('login_and_register.already_have_account') }}<br>
              <BcLink to="/login" :target="Target.Internal" class="link">
                {{ $t('login_and_register.login_here') }}
              </BcLink>
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

.content {
  display: flex;
  flex-direction: column;

  .caption {
    position: relative;
    margin-top: 20px;
    margin-left: auto;
    margin-right: auto;
    .title {
      font-size: 26px;
      margin-bottom: 8px;
    }
    .purpose {
      font-size: 20px;
    }
  }

  .container {
    position: relative;
    width: 355px;
    height: 420px;
    margin: auto;
    margin-top: 30px;
    margin-bottom: 30px;
    padding: var(--padding-large);
    box-sizing: border-box;

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
          margin-bottom: 8px;
        }
        .agreement {
          display: flex;
          flex-direction: row;
          gap: 10px;
          .checkbox {
            margin-top: auto;
            margin-bottom: auto;
            :deep(.p-checkbox-box) {
              width: 30px;
              height: 30px;
            }
          }
        }
      }

      .last-row {
        position: relative;
        display: flex;
        margin-top: auto;

        .login-invitation {
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
}
</style>
