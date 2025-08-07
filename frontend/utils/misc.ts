export const addUpValues = (obj?: Record<string, number>): number => {
  if (!obj) {
    return 0
  }
  return Object.values(obj).reduce((sum, val) => sum + val, 0)
}
