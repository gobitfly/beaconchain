export default defineEventHandler((event) => {
  return redirectToV1(event, '/register')
})
