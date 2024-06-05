import { intersection } from 'lodash-es'
import type { SwipeCallback, SwipeDirection, SwipeOptions } from '~/types/swipe'

export const useSwipe = (options: Ref<SwipeOptions> = ref({
  directinoal_threshold: 100,
  directions: ['all']
}), bounce = true) => {
  const touchStartX = ref(0)
  const touchEndX = ref(0)
  const touchStartY = ref(0)
  const touchEndY = ref(0)
  const touchableElement = ref<HTMLElement | undefined>()

  const onSwipe = ref<SwipeCallback>() // triggers if any swipe happend

  const onTouchStart = (event: TouchEvent) => {
    touchStartX.value = event.changedTouches[0].screenX
    touchStartY.value = event.changedTouches[0].screenY
  }
  const onTouchEnd = (event: TouchEvent) => {
    touchEndX.value = event.changedTouches[0].screenX
    touchEndY.value = event.changedTouches[0].screenY

    if (!handleGesture(event) && touchableElement.value) {
      touchableElement.value.style.transform = ''
    }
  }

  const onTouchMove = (event: TouchEvent) => {
    if (!bounce || !touchableElement.value) {
      return
    }
    let divX = event.changedTouches[0].screenX - touchStartX.value
    const directions = options.value.directions ?? []
    if (!intersection(directions, ['all', 'left']).length && divX < 0) {
      divX = 0
    }
    if (!intersection(directions, ['all', 'right']).length && divX > 0) {
      divX = 0
    }
    let divY = event.changedTouches[0].screenY - touchStartY.value
    if (!intersection(directions, ['all', 'top']).length && divY < 0) {
      divY = 0
    }
    if (!intersection(directions, ['all', 'bottom']).length && divY > 0) {
      divY = 0
    }
    const transform = `translate(${divX}px, ${divY}px)`
    touchableElement.value.style.transform = transform
  }

  const handleGesture = (event: TouchEvent) => {
    const divX = Math.abs(touchEndX.value - touchStartX.value)
    const divY = Math.abs(touchEndY.value - touchStartY.value)
    const threshold = options.value?.directinoal_threshold ?? 0
    const gDirections: SwipeDirection[] = []
    if (touchEndX.value < touchStartX.value && divX > threshold) {
      gDirections.push('left')
    }

    if (touchEndX.value > touchStartX.value && divX > threshold) {
      gDirections.push('right')
    }

    if (touchEndY.value < touchStartY.value && divY > threshold) {
      gDirections.push('top')
    }

    if (touchEndY.value > touchStartY.value && divY > threshold) {
      gDirections.push('bottom')
    }
    if (gDirections.length) {
      gDirections.push('all')
    }

    if (intersection(gDirections, options.value.directions).length && onSwipe.value?.(event, gDirections)) {
      return true
    }
  }

  const setElement = (elem: HTMLElement, callback: SwipeCallback) => {
    clearElement()
    touchableElement.value = elem
    onSwipe.value = callback
    if (touchableElement.value) {
      touchableElement.value.addEventListener('touchstart', onTouchStart, false)
      touchableElement.value.addEventListener('touchend', onTouchEnd, false)
      touchableElement.value.addEventListener('touchcancel', onTouchEnd, false)
      touchableElement.value.addEventListener('touchmove', onTouchMove, false)
    }
  }

  const clearElement = () => {
    if (touchableElement.value) {
      touchableElement.value.removeEventListener('touchstart', onTouchStart)
      touchableElement.value.removeEventListener('touchend', onTouchEnd)
      touchableElement.value.addEventListener('touchcancel', onTouchEnd, false)
      touchableElement.value.removeEventListener('touchmove', onTouchMove)
      touchableElement.value = undefined
    }
  }

  onUnmounted(() => {
    clearElement()
  })

  return {
    setTouchableElement: (elem: HTMLElement, callback: SwipeCallback) => setElement(elem, callback)
  }
}
