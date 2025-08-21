<script lang="ts" setup>
const { isLoggedIn } = useUserStore()

useBcSeo('user_settings.title')
const { stripeInit } = useStripeProvider()
const {
  getProducts,
  products,
} = useProductsStore()

const buttonsDisabled = ref(false)

await useAsyncData('get_products', () => getProducts())
const { navigateToV1Login } = useV1Login()
watch(
  products,
  () => {
    if (products.value?.stripe_public_key) {
      stripeInit(products.value.stripe_public_key)
    }
  },
  { immediate: true },
)

if (!isLoggedIn.value) {
  // only users that are logged in can view this page
  // TODO: This should maybe be part of the middleware
  await navigateToV1Login()
}
</script>

<template>
  <BcPageWrapper>
    <div class="settings-container">
      <UserSettingsSubscriptions v-model="buttonsDisabled" />
      <UserSettingsEmail v-model="buttonsDisabled" />
      <UserSettingsPassword v-model="buttonsDisabled" />
      <UserSettingsDeleteAccount v-model="buttonsDisabled" />
    </div>
  </BcPageWrapper>
</template>

<style lang="scss" scoped>
.settings-container {
  position: relative;
  margin-left: auto;
  margin-right: auto;
  max-width: 750px;

  display: flex;
  flex-direction: column;
  gap: var(--padding);
}
</style>
