<script setup lang="ts">
// import { object as yupObject } from 'yup'
// import { useForm } from 'vee-validate'
// import { useUserStore } from '~/stores/useUserStore'
// import { Target } from '~/types/links'
// import {
//   handleMobileAuth, provideMobileAuthParams,
// } from '~/utils/mobileAuth'

// const { t: $t } = useTranslation()
// const { doLogin } = useUserStore()
// const toast = useBcToast()
// const route = useRoute()
// const { promoCode } = usePromoCode()

if (!useRuntimeConfig().public.showInDevelopment) navigateTo('https://beaconcha.in/requestReset', { external: true })
// useBcSeo('auth.login_and_register.title_login')

// const {
//   defineField, errors, handleSubmit,
// } = useForm({
//   validationSchema: yupObject({
//     email: emailValidation($t),
//     password: passwordValidation($t),
//   }),
// })

// const [
//   email,
//   emailAttrs,
// ] = defineField('email')
// const [
//   password,
//   passwordAttrs,
// ] = defineField('password')

// const onSubmit = handleSubmit(async (values) => {
//   try {
//     await doLogin(values.email, values.password)

//     if (handleMobileAuth(route.query)) {
//       return
//     }

//     if (promoCode) {
//       await navigateTo({
//         path: '/pricing', query: { promoCode },
//       })
//     }
//     else {
//       await navigateTo('/')
//     }
//   }
//   catch (error) {
//     password.value = ''
//     toast.showError({
//       detail: $t('auth.login_and_register.error_login_message'),
//       group: $t('auth.login_and_register.error_login_group'),
//       summary: $t('auth.login_and_register.error_title'),
//     })
//   }
// })

// const canSubmit = computed(() => email.value && password.value && !Object.keys(errors.value).length)

// const registerLink = computed(() => {
//   return provideMobileAuthParams(route.query, '/register')
// })
const { t: $t } = useTranslation()
const input = ref('')
const fieldName = 'email'
const {
  handleSubmit,
} = useForm<{ [fieldName]: string }>()
const onSubmit = handleSubmit((values) => {
  console.log('submit', values)
})
</script>

<template>
  <BcPageWrapper :minimalist-header="true">
    <BaseLayoutAuth>
      <BaseCard
        heading-tag="h1"
        :heading="$t('auth.request_reset.title')"
      >
        <BaseGutter>
          {{ $t('auth.request_reset.message') }}
          <BaseForm @submit="onSubmit">
            <BaseFormInputEmail
              v-model="input"
              :field-name
              :label="false"
              autofocus
            />
          </BaseForm>
        </BaseGutter>
        <template #footer>
          <BaseButton variant="secondary">
            Back to Login
          </BaseButton>
          <BaseButton @click="handleSubmit">
            Send Link
          </BaseButton>
        </template>
      </BaseCard>
    </BaseLayoutAuth>
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
