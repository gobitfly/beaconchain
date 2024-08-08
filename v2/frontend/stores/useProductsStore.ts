import { defineStore } from 'pinia'
import type {
  InternalGetProductSummaryResponse,
  ProductSummary,
} from '~/types/api/user'
import { API_PATH } from '~/types/customFetch'
import {
  ProductCategoryPremium,
  ProductStoreAndroidPlaystore,
  ProductStoreIosAppstore,
} from '~/types/api/user'

const productsStore = defineStore('products_store', () => {
  const data = ref<ProductSummary>()

  return { data }
})

export function useProductsStore() {
  const { fetch } = useCustomFetch()
  const { data } = storeToRefs(productsStore())
  const { user } = useUserStore()

  const products = computed(() => data.value)

  const bestPremiumProduct = computed(() => {
    return data.value?.premium_products.reduce(
      (max, product) =>
        product.price_per_year_eur > max.price_per_year_eur ? product : max,
      data.value.premium_products[0],
    )
  })

  const currentPremiumSubscription = computed(() => {
    return user.value?.subscriptions?.find(
      sub => sub.product_category === ProductCategoryPremium,
    )
  })

  const isPremiumSubscribedViaApp = computed(() => {
    const store = currentPremiumSubscription.value?.product_store
    return (
      store === ProductStoreAndroidPlaystore
      || store === ProductStoreIosAppstore
    )
  })

  async function getProducts() {
    if (data.value) {
      return data.value
    }

    const res = await fetch<InternalGetProductSummaryResponse>(
      API_PATH.PRODUCT_SUMMARY,
    )

    data.value = res.data
    return res
  }

  return {
    bestPremiumProduct,
    currentPremiumSubscription,
    getProducts,
    isPremiumSubscribedViaApp,
    products,
  }
}
