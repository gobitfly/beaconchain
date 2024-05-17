import { defineStore } from 'pinia'
import type { InternalGetProductSummaryResponse } from '~/types/api/user'
import { API_PATH } from '~/types/customFetch'

const productsStore = defineStore('products_store', () => {
  const data = ref < InternalGetProductSummaryResponse>()

  return { data }
})

export function useProductsStore () {
  const { fetch } = useCustomFetch()
  const { data } = storeToRefs(productsStore())

  const products = computed(() => data.value)

  async function getProducts () {
    if (data.value) {
      return data.value
    }

    const res = await fetch<InternalGetProductSummaryResponse>(API_PATH.PRODUCT_SUMMARY)

    data.value = res
    return res
  }

  return { products, getProducts }
}
