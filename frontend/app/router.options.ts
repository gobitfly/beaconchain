export default {
  scrollBehavior(
    _to: any,
    _from: any,
    savedPosition: { left: number, top: number } | null,
  ) {
    return { _to, ...savedPosition }
  },
}
