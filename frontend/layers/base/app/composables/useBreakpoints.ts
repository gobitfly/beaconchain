import {
  breakpointsTailwind,
  useBreakpoints as useBreakpointsVueuse,
} from '@vueuse/core'

export const useBreakpoints = () => {
  return useBreakpointsVueuse(breakpointsTailwind)
}
