<script setup lang="ts">
import { useField, useForm } from 'vee-validate'
import { Target } from '~/types/links'
import { tOf } from '~/utils/translation'
import { API_PATH } from '~/types/customFetch'
import { setTranslator, validateAddress, validatePassword, validateAgreement } from '~/utils/userValidation'

const { t: $t } = useI18n()
const { fetch } = useCustomFetch()
const toast = useBcToast()

useBcSeo('login_and_register.title_register')

const { handleSubmit, errors } = useForm()
const { value: email } = useField<string>('email', validateAddress)
const { value: password } = useField<string>('password', validatePassword)
const { value: passwordConfirm } = useField<string>('passwordConfirm', validatePasswordConfirmation)
const { value: agreement } = useField<boolean>('agreement', validateAgreement)

setTranslator($t)

function validatePasswordConfirmation (value: string) : true|string {
  if (!value) {
    return $t('login_and_register.retype_password')
  }
  if (value !== password.value) {
    return $t('login_and_register.passwords_dont_match')
  }
  return true
}

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

const canSubmit = computed(() => email.value && password.value && passwordConfirm.value && agreement.value && !Object.keys(errors.value).length)
const addressError = ref<string|undefined>(undefined)
const passwordError = ref<string|undefined>(undefined)
const passwordConfirmError = ref<string|undefined>(undefined)
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
            <label for="password" class="label">{{ $t('login_and_register.choose_password') }}</label>
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
          <div class="input-row">
            <label for="passwordConfirm" class="label">{{ $t('login_and_register.confirm_password') }}</label>
            <InputText
              id="passwordConfirm"
              v-model="passwordConfirm"
              type="password"
              :class="{ 'p-invalid': errors?.passwordConfirm }"
              aria-describedby="text-error"
              @focus="passwordConfirmError = undefined"
              @blur="passwordConfirmError = errors?.passwordConfirm"
            />
            <div class="p-error">
              {{ passwordConfirmError || '&nbsp;' }}
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
        @include fonts.small_text;
      }
    }
  }
}
</style>
