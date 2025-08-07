import { LINK } from '~/utils/externalLinks'

export default defineEventHandler((event) => {
  return sendRedirect(event, LINK.shop, 307)
})
