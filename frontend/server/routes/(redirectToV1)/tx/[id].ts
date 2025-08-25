export default defineEventHandler((event) => {
  const id = getRouterParam(event, 'id')
  return redirectToV1(event, `/tx/${id}`)
})
