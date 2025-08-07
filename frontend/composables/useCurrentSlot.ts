export const useCurrentSlot = () => {
  const latestState = useFetchedData('/api/latest-state')
  const currentSlot = computed(() => latestState.value?.current_slot ?? 0)
  return currentSlot
}
