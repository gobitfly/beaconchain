export const useProduct = () => {
  const data = useFetchedData('productSummary')
  const hasData = computed(() => Object.keys(data.value ?? {}).length > 0)
  const { $api } = useNuxtApp()
  if (!hasData.value) {
    callOnce(async () => {
      data.value = await $api('/api/bff/product-summary')
    })
  }

  const addons = computed(() =>
    Object.fromEntries(data.value?.extra_dashboard_validators_premium_addons?.map(addon => [
      addon.product_name,
      addon,
    ]) ?? []))
  const addonEffectiveBalanceIncrease32k = computed(() => addons.value['32000 effective balance increase per dashboard'])
  const addonEffectiveBalanceIncrease320k = computed(() => addons.value['320000 effective balance increase per dashboard'])

  const effectiveBalancePerDashboardLimit = computed(() => data.value?.effective_balance_per_dashboard_limit ?? 0)

  const products = computed(() =>
    Object.fromEntries(data.value?.premium_products?.map(product => [
      product.product_name,
      product,
    ]) ?? []))
  const productDolphin = computed(() => products.value['Dolphin'])
  const productFree = computed(() => products.value['Free'])
  const productGuppy = computed(() => products.value['Guppy'])
  const productOrca = computed(() => products.value['Orca'])

  return {
    addonEffectiveBalanceIncrease32k,
    addonEffectiveBalanceIncrease320k,
    // effective_balance_per_dashboard_limit: ,
    effectiveBalancePerDashboardLimit,
    productDolphin,
    // stripe_public_key: data.value?.stripe_public_key,
    productFree,
    productGuppy,
    productOrca,
    products,
  }
}
