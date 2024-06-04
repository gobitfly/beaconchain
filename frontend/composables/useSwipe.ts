export type SwipeCallback = (event: TouchEvent) => void;
export type SwipeOptions = {
    directinoal_threshold?: number; // Pixels offset to trigger swipe
};
export const useSwipe = (options: Ref<SwipeOptions> = ref({
  directinoal_threshold: 100
}), bounce = true) => {
  const touchStartX = ref(0)
  const touchEndX = ref(0)
  const touchStartY = ref(0)
  const touchEndY = ref(0)
  const touchableElement = ref<HTMLElement | undefined>()

  const onSwipeLeft = ref<SwipeCallback>()
  const onSwipeRight = ref<SwipeCallback>()
  const onSwipeUp = ref<SwipeCallback>()
  const onSwipeDown = ref<SwipeCallback>()

  const onTouchStart = (event: TouchEvent) => {
    touchStartX.value = event.changedTouches[0].screenX
    touchStartY.value = event.changedTouches[0].screenY
  }
  const onTouchEnd = (event: TouchEvent) => {
    touchEndX.value = event.changedTouches[0].screenX
    touchEndY.value = event.changedTouches[0].screenY
    handleGesture(event)
    if (touchableElement.value) {
      touchableElement.value.style.transform = ''
    }
  }

  const onTouchMove = (event: TouchEvent) => {
    if (!bounce || !touchableElement.value) {
      return
    }
    let divX = event.changedTouches[0].screenX - touchStartX.value
    if (!onSwipeLeft.value && divX < 0) {
      divX = 0
    }
    if (!onSwipeRight.value && divX > 0) {
      divX = 0
    }
    let divY = event.changedTouches[0].screenY - touchStartY.value
    if (!onSwipeUp.value && divY < 0) {
      divY = 0
    }
    if (!onSwipeDown.value && divY > 0) {
      divY = 0
    }
    const transform = `translate(${divX}px, ${divY}px)`
    touchableElement.value.style.transform = transform
  }

  const handleGesture = (event: TouchEvent) => {
    const divX = Math.abs(touchEndX.value - touchStartX.value)
    const divY = Math.abs(touchEndY.value - touchStartY.value)
    if (touchEndX.value < touchStartX.value && divX > (options.value?.directinoal_threshold ?? 0)) {
      onSwipeLeft.value?.(event)
    }

    if (touchEndX.value > touchStartX.value && divX > (options.value?.directinoal_threshold ?? 0)) {
      onSwipeRight.value?.(event)
    }

    if (touchEndY.value < touchStartY.value && divY > (options.value?.directinoal_threshold ?? 0)) {
      onSwipeUp.value?.(event)
    }

    if (touchEndY.value > touchStartY.value && divY > (options.value?.directinoal_threshold ?? 0)) {
      onSwipeDown.value?.(event)
    }
  }

  const setElement = (elem: HTMLElement) => {
    clearElement()
    touchableElement.value = elem
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
    setTouchableElement: (elem: HTMLElement) => setElement(elem),
    onSwipeLeft: (callback: SwipeCallback) => (onSwipeLeft.value = callback),
    onSwipeRight: (callback: SwipeCallback) => (onSwipeRight.value = callback),
    onSwipeUp: (callback: SwipeCallback) => (onSwipeUp.value = callback),
    onSwipeDown: (callback: SwipeCallback) => (onSwipeDown.value = callback)
  }
}
