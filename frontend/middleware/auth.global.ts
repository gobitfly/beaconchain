export default defineNuxtRouteMiddleware((to) => {
  const { isLoggedIn } = useUserStore()
  const isLoginRoute = to.name === 'login'

  if (!isLoginRoute && isClientSide) {
    const currentUrl = useCookie('currentUrl')
    currentUrl.value = to.fullPath
  }

  // console.log('route name: ', to.name, isLoggedIn.value)
  // if (to.path === '/notifications' && !isLoggedIn) {
  //   console.log('from middleware', useUserStore())
  //   return '/login'
  // }
})
