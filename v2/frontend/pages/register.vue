<script setup lang="ts">
import { useField, useForm } from 'vee-validate'
import { useUserStore } from '~/stores/useUserStore'
import { REGEXP_VALID_EMAIL } from '~/utils/regexp'
import { Target } from '~/types/links'
import { tOf } from '~/utils/translation'

const { t: $t } = useI18n()
const { doLogin } = useUserStore()
const toast = useBcToast()

useBcSeo('login_and_register.title_register')

const { handleSubmit, errors } = useForm()
const { value: email } = useField<string>('email', validateAddress)
const { value: password1 } = useField<string>('password1', validatePassword1)
const { value: password2 } = useField<string>('password2', validatePassword2)
const { value: agreement } = useField<boolean>('agreement', validateAgreement)

function validateAddress (value: string) : true|string {
  if (!value) {
    return $t('login_and_register.no_email')
  }
  if (!REGEXP_VALID_EMAIL.test(value)) {
    return $t('login_and_register.invalid_email')
  }
  return true
}

function validatePassword1 (value: string) : true|string {
  if (!value) {
    return $t('login_and_register.no_password')
  }
  if (value.length < 5) {
    return $t('login_and_register.invalid_password')
  }
  return true
}

function validatePassword2 (value: string) : true|string {
  if (!value) {
    return $t('login_and_register.retype_password')
  }
  if (value !== password1.value) {
    return $t('login_and_register.passwords_dont_match')
  }
  return true
}

function validateAgreement (value : boolean) : true|string {
  if (!value) {
    return $t('login_and_register.not_agreed')
  }
  return true
}

const onSubmit = handleSubmit(async (values) => {
  try {
    await doLogin(values.email, values.password)
    await navigateTo('/')
  } catch (error) {
    toast.showError({ summary: $t('login_and_register.error_toast_title'), group: $t('login_and_register.error_toast_group'), detail: $t('login_and_register.error_toast_message') })
  }
})

const canSubmit = computed(() => email.value && password1.value && password2.value && agreement.value && !Object.keys(errors.value).length)
const addressError = ref<string|undefined>(undefined)
const password1Error = ref<string|undefined>(undefined)
const password2Error = ref<string|undefined>(undefined)
const agreementError = ref<string|undefined>(undefined)
</script>

<template>
  <BcPageWrapper>
    <div class="content">
      <div class="caption">
        <div class="title">
          Sign up to beaconcha.in
        </div>
        <div class="purpose">
          to manage and monitor your validators.
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
            <label for="password1" class="label">{{ $t('login_and_register.choose_password') }}</label>
            <InputText
              id="password1"
              v-model="password1"
              type="password"
              :class="{ 'p-invalid': errors?.password1 }"
              aria-describedby="text-error"
              @focus="password1Error = undefined"
              @blur="password1Error = errors?.password1"
            />
            <div class="p-error">
              {{ password1Error || '&nbsp;' }}
            </div>
          </div>
          <div class="input-row">
            <label for="password2" class="label">{{ $t('login_and_register.confirm_password') }}</label>
            <InputText
              id="password2"
              v-model="password2"
              type="password"
              :class="{ 'p-invalid': errors?.password2 }"
              aria-describedby="text-error"
              @focus="password2Error = undefined"
              @blur="password2Error = errors?.password2"
            />
            <div class="p-error">
              {{ password2Error || '&nbsp;' }}
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
    }
    .purpose {
      font-size: 20px;
    }
  }

  .container {
    position: relative;
    width: 355px;
    height: 400px;
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
