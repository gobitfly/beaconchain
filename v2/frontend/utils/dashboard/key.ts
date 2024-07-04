export function isPublicDashboardKey (value?: string): boolean {
  if (!value) {
    return true
  }

  const id = parseInt(value)
  return isNaN(id)
}

export function isSharedKey (value?: string): boolean {
  return !!value?.startsWith('v-')
}
