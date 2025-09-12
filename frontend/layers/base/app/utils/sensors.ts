import { onClickOutside as onClickOutsideVueuse } from '@vueuse/core'

export const onClickOutside = (target: Ref<HTMLElement>, callback: (event: PointerEvent) => void) => {
  onClickOutsideVueuse(target, callback)
}
