export const useV1Domain = () => {
  return useRuntimeConfig().public.v1Domain || 'https://beaconcha.in'
}
