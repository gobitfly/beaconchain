<script lang="ts" setup>
import { ProductStoreAndroidPlaystore, ProductStoreIosAppstore } from '~/types/api/user'

const { t: $t } = useI18n()
const { currentPremiumSubscription } = useProductsStore()

const visible = computed(() => {
  const store = currentPremiumSubscription.value?.product_store
  return (store === ProductStoreAndroidPlaystore || store === ProductStoreIosAppstore)
})
</script>

<template>
  <div v-if="visible" class="banner">
    {{ $t('pricing.premium_via_app_banner') }}
  </div>
</template>

<style lang="scss" scoped>
.banner {
  width: 100%;
  max-width: var(--pricing-content-width);
  box-sizing: border-box;
  border: 2px solid var(--primary-color);
  border-radius: 7px;
  color: var(--text-color);
  background-color: var(--container-background);
  padding: 10px 30px;

  margin-bottom: 55px;
}
</style>
