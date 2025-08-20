<script setup lang="ts">
import type { NuxtError } from '#app'
import { ERROR_CODE } from '~/shared/utils/helper'

defineProps({
  error: Object as () => NuxtError,
})

// onBeforeUnmount(() => {
//   clearError()
// })
const { t: $t } = useTranslation()
// const v1Domain = useDomain('v1')
const loginUrl = useLoginUrl()
</script>

<template>
  <div>
    <NuxtLayout name="default">
      <div class="error-page">
        <BcError
          v-if="error?.statusMessage === ERROR_CODE.GUEST_DASHBOARD_ID_INVALID"
          description="dashboard.error.400.description"
          title="dashboard.error.400.title"
        />

        <BcError
          v-else-if="error?.statusMessage === ERROR_CODE.EFFECTIVE_BALANCE_EXCEEDS_LIMIT"
          description="dashboard.validator.management.validators_limit_exceeded"
          title="dashboard.error.400.title"
        >
          <BcLink
            to="/pricing"
            class="link"
          >
            {{ $t('error.read_more_about_how_to_upgrade') }}
          </BcLink>
        </BcError>

        <BcError
          v-else-if="error?.statusMessage === ERROR_CODE.UNAUTHORIZED_DASHBOARD"
          description="dashboard.error.401.description"
          title="dashboard.error.401.title"
        >
          <BcLink
            :to="loginUrl"
            class="link"
          >
            {{ $t('error.go_to_login_page') }}
          </BcLink>
        </BcError>

        <BcError
          v-else-if="error?.statusCode === 404"
          description="error.404.description"
          title="error.404.title"
        />

        <template v-else>
          <h1>{{ $t('error.we_are_sorry_error_occurred') }} ({{ error?.statusCode }})</h1>
          <p v-if="error?.statusMessage">
            {{ error.statusMessage }}
          </p>
        </template>

        <section>
          <BcLink
            to="/"
            class="link"
          >
            {{ $t('navigation.back_to_home') }}
          </BcLink>
        </section>
      </div>
    </NuxtLayout>
  </div>
</template>

<style lang="scss" scoped>
.error-page {
  display: flex;
  gap: 1rem;
  flex-direction: column;
  justify-content: center;
  align-items: center;
}
</style>
