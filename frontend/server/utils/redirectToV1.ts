import type { H3Event } from 'h3'

const v1Domain = useRuntimeConfig().public.v1Domain ?? 'https://beaconcha.in'
export const redirectToV1 = async (event: H3Event, path: `/${string}`) => {
  return sendRedirect(event, `${v1Domain}${path}`, 307)
}
