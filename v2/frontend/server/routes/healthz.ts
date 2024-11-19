export default defineEventHandler(() => {
  // this endpoint is used for the `loadbalancer` to check if the server is up (status 200)
  return {
    status: 'ok',
  }
})
