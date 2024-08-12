import { defineStore } from 'pinia'
import { getCSRFHeader } from '~/utils/fetch'

/**
  The csrf header is added to non GET requests to prevent Cross-site request forgery
  We get the csrf header from GET requests put them in this store and apply them to non GET requests.
**/

const csrfStore = defineStore('csrf_store', () => {
  const header = ref<[string, string] | null | undefined>()
  return { header }
})

export function useCsrfStore() {
  const { header } = storeToRefs(csrfStore())

  const csrfHeader = computed(() => header.value)

  function setCsrfHeader(headers: Headers) {
    const h = getCSRFHeader(headers)
    if (h) {
      header.value = h
    }
  }

  return {
    csrfHeader,
    setCsrfHeader,
  }
}
