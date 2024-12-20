export const addUpValues = (obj?: Record<string, number>): number => {
  if (!obj) {
    return 0
  }
  return Object.values(obj).reduce((sum, val) => sum + val, 0)
}

export function isInt(value?: string): boolean {
  if (!value) {
    return false
  }
  const parsed = parseInt(value)
  return !isNaN(parsed) && `${parsed}` === value
}
