export function isGuestDashboardKey(value?: string): boolean {
  if (!value) {
    return true
  }
  if (value.startsWith('v-')) {
    return false
  }

  const id = parseInt(value)
  return isNaN(id)
}

export function isSharedDashboardKey(value?: string): boolean {
  return !!value?.startsWith('v-')
}
