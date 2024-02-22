// For Nuxt 3
import { config } from '@fortawesome/fontawesome-svg-core'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
// import { fas } from '@fortawesome/pro-solid-svg-icons'
// import { far } from '@fortawesome/pro-regular-svg-icons'
// import { fab } from '@fortawesome/free-brands-svg-icons'

// This is important, we are going to let Nuxt worry about the CSS
config.autoAddCss = false

// You can add your icons directly in this plugin. See other examples for how you
// can add other styles or just individual icons.
/*
If we want to include all icons we could add them like this, but it would be better for tree shaking to import them one by one
library.add(fas)
library.add(far)
library.add(fab)
*/

export default defineNuxtPlugin((nuxtApp) => {
  nuxtApp.vueApp.component('font-awesome-icon', FontAwesomeIcon, {
    css: [
      '@fortawesome/fontawesome-svg-core/styles.css'
    ]
  })
})
