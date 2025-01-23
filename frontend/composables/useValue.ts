export function useValue() {
  const converter = computed(() => {
    const weiToValue = (
    ) => ({
      fullLabel: 'zzzzzz',
      label: 'aaaaaaa',
    })
    return { weiToValue }
  })

  return { converter }
}
