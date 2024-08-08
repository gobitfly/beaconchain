export type SwipeDirection = 'all' | 'bottom' | 'left' | 'right' | 'top'
// if callback returns true we keep the element at it's position
// (example: dialog hides where you left it and not pops back)
export type SwipeCallback = (
  event: TouchEvent,
  directions: SwipeDirection[],
) => boolean | undefined
export type SwipeOptions = {
  directional_threshold?: number // Pixels offset to trigger swipe
  directions?: SwipeDirection[]
}
