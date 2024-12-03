export default {
  scrollBehavior(
    _to: any,
    _from: any,
    savedPosition: null | {
      left: number,
      top: number,
    },
  ) {
    return {
      _to,
      ...savedPosition,
    }
  },
}
