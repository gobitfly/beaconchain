export type SwipeDirection = 'left' | 'top' | 'right' | 'bottom' | 'all'
export type SwipeCallback = (event: TouchEvent, directions: SwipeDirection[]) => void | boolean; // if callback returns true we keep the element at it's position (example: dialog hides where you left it and not pops back)
export type SwipeOptions = {
    directinoal_threshold?: number; // Pixels offset to trigger swipe
    directions?: SwipeDirection[];
    invalidSwipeClasses?: string[];
};
