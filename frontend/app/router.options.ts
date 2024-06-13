export default {
  scrollBehavior (_to: any, _from: any, savedPosition: { left: number, top: number } | null) {
    if (savedPosition) {
      return savedPosition
    } else {
      return { top: 0 }
    }
  }
}
