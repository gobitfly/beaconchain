export function isGuestDashboardKey(value?: string): boolean {
  if (!value) {
    return true
  }

  const id = parseInt(value)
  return isNaN(id)
}

export function isSharedDashboardKey(value?: string): boolean {
  return !!value?.startsWith('v-')
}
