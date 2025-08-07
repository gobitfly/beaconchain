export default defineEventHandler((event) => {
  // const result = getSession(event, { password: 'test' })
  // console.log('result', result)
  // clearSession(event, { })
  const query = getQuery(event)
  return query
})
