<script lang="ts" setup>
const { isLoggedIn } = useUserStore()

useBcSeo('user_settings.title')
const { stripeInit } = useStripeProvider()
const { products, getProducts } = useProductsStore()

const buttonsDisabled = ref(false)

await useAsyncData('get_products', () => getProducts())
watch(products, () => {
  if (products.value?.stripe_public_key) {
    stripeInit(products.value.stripe_public_key)
  }
}, { immediate: true })

// TODO: This seems to be triggered to soon
if (!isLoggedIn.value) {
  // only users that are logged in can view this page
  await navigateTo('/login')
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
  width: 750px;

  display: flex;
  flex-direction: column;
  gap: var(--padding);
}
</style>
