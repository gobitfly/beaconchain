import { defineStore } from 'pinia'

/**
  The csrf header is added to non GET requests to prevent Cross-site request forgery
  We get the csrf header from GET requests put them in this store and apply them to non GET requests.
**/

const csrfStore = defineStore('csrf_store', () => {
  const tokenCsrf = ref<string >('')
  return { tokenCsrf }
})

export function useCsrfStore() {
  const { tokenCsrf } = storeToRefs(csrfStore())

  function setTokenCsrf(token: string) {
    tokenCsrf.value = token
  }

  return {
    setTokenCsrf,
    tokenCsrf,
  }
}
