<!-- eslint-disable vue/max-len -- TODO: plz fix this -->
<script setup lang="ts">
import {
  type ComponentPublicInstance, warn,
} from 'vue'

const DEBUG = false // Use Chromium or Chrome. Firefox will show messages with broken indentation, illegible codes and no color differenciating the types of the messages.

const ResizeObserverLagMargin = 1 // This safety margin is important, because the resizing observer happens to lag. If a small decrease of width making the frame as large as its content does not trigger the observer, then it will not fire anymore because the frame cannot shrink anymore.

const props = defineProps<{
  class?: string | string[], // to make the list of classes reactive
  ellipses?: number | number[], // If number: number of ellipses to use (the same for any room available), 1 by default. If array, its meaning is: [room above which two `…` are used, room above which three `…` are used, and so on]. Ex: [8,30,100] tells the component to use one ellipsis if there is room for 8 characters or less, or two ellipses between 9 and 30 characters, and so on.
  initialFlexGrow?: number, // If the component has no defined size (meaning that its width collapses to 0 when it contains nothing) then you must set a value in this props.
  // !! The props below are for internal use only !!
  meCallbackToInformParentAboutChanges?: typeof enterUpdateCycleAsAparent,
  text?: string,
  widthMediaqueryThreshold?: number, // Very important: if a `@media (min-width: AAApx)` or a `@media (max-width: AAApx)` somewhere in your CSS has an effect on the size of the component (sudden changes of width), give AAA to this pros.
}>()

interface ExposedMembers {
  amIofDefinedWidth: ComputedRef<boolean>,
  enterUpdateCycleAsAparent: typeof enterUpdateCycleAsAparent,
  getReadyForUpdate: typeof getReadyForUpdate,
  howMuchCanIshrinkOrGrow: typeof howMuchCanIshrinkOrGrow,
  saveFinalState: typeof saveFinalState,
  settleAfterUpdate: typeof settleAfterUpdate,
  updateContent: typeof updateContent,
  whatIsMyFlexGrow: typeof whatIsMyFlexGrow,
}

interface MiddleEllipsis extends ComponentPublicInstance, ExposedMembers {}

enum WhatIcanBe {
  Error = 0,
  Parent,
  Child,
  Standalone,
}

enum WidthMode {
  NoFlexGrow,
  NoWidth,
  FixedFlexGrow,
  FixedWidth,
}

// This identifiers will tell us how the gap between the end of the text and the right edge of the frame changes
enum UpdateReason {
  None = 0,
  GapChangeToBeDetermined,
  GapChangeMinus,
  GapChangePlus,
}

enum SignalDirection {
  ChildrenToParent,
  ParentToChildren,
}

type TextProperties = { text: string,
  width: number, }

const _s = useSlots() // Not meant to be used directly. Use the reactive variable `slot` defined just below:
const slot = computed(() => (_s.default ? _s.default() : [])) // `slot`s is always an array, empty if there is no slot

const innerElements = {
  allInstanciatedElements: ref<(ComponentPublicInstance | MiddleEllipsis)[]>(
    [],
  ), // Instanciated elements from our slot. This array is filled by Vue in the <template>.
  isAnUpdateOrdered: true,
  slotNonce: 0,
  // The following arrays are filled by us, each time the slot is modified:
  widthDefinedChildren: [] as MiddleEllipsis[], // List of instanciated elements from our slot that are MiddleEllipsis children with a defined width.
  widthUndefinedChildren: [] as MiddleEllipsis[], // List of instanciated elements from our slot that are MiddleEllipsis children with an undefined width.
}
const frameSpan = ref<HTMLSpanElement>(null as unknown as HTMLSpanElement)
let frameStyle: CSSStyleDeclaration
let frameText = props.text || '' // After mounting, this variable will always mirror the text in the frame. Before mounting, it contains the full text so that <template> can display it during isServer.

let mediaqueryWidthListener: MediaQueryList
let delayedForcedUpdateIncoming = false

let classPropsDuringLastUpdate = props.class || ''
let textPropsDuringLastUpdate = props.text || ''
let initialFlexGrowDuringLastUpdate: number | undefined
let ellipsesPropsDuringLastUpdate: number | number[] | undefined = 1
let textAfterLastUpdate: TextProperties = {
  text: '',
  width: 0,
}
let widthAvailableDuringLastUpdate = 0 // used by determineReason() to find out why an update is needed, during the update process
let frameWidthAfterLastUpdate = 0 // used by determineReason() to find out why an update is needed, outside the update process
let lastMeasuredFrameWidth = 0
let currentAdditionalWidthAvailable = 0
let currentText = ''
const canvasContextToCalculateTextWidths = (
  isServer ? null : document.createElement('canvas').getContext('2d')
) as CanvasRenderingContext2D
const lastTextWidthCalculation: TextProperties = {
  text: '',
  width: 0,
}
let amImounted = false
let didTheResizingObserverFireSinceMount = false
let amIreadyForUpdate = false // our parent can call function getReadyForUpdate() as we can, so we use this variable to prevent multiple executions of it in a row
let lastSlotNonceWhenChecked = -1

let numberOfClippings = 0
let totalIterationsWhenClipping = 0

const amIofDefinedWidth = computed(() => {
  // TODO: Maybe check whether the width is defined in the CSS of the component if-and-only-if props.initialFlexGrow is not set.
  //       Problem if done: it would be a slow operation at execution time just to provide a security against the programmer during development (because an inconsistency here causes unwanted results on the screen anyway)
  return !props.initialFlexGrow
})

const doIobserveMyResizing = computed(() => {
  return amIofDefinedWidth.value
})

const amIinsideAparent = computed(() => {
  return !!props.meCallbackToInformParentAboutChanges
})

const whatIam = computed(() => {
  if (!amIofDefinedWidth.value && !amIinsideAparent.value) {
    warn(
      'MiddleEllipsis cannot have an undefined width because it is outside a MiddleEllipsis container.',
    )
    return WhatIcanBe.Error
  }
  if (slot.value[0]) {
    if (props.text) {
      warn(
        'MiddleEllipsis cannot receive any content in props `text` because it is a container.',
      )
      return WhatIcanBe.Error
    }
    return WhatIcanBe.Parent // Parents in parents are considered Parents because slot.value[0] has been tested before the following:
  }
  if (props.text === undefined) {
    warn('MiddleEllipsis expects its props `text` to be set.')
    return WhatIcanBe.Error
  }
  return amIinsideAparent.value ? WhatIcanBe.Child : WhatIcanBe.Standalone
})

const exposedMembers: ExposedMembers = {
  amIofDefinedWidth,
  enterUpdateCycleAsAparent,
  getReadyForUpdate,
  howMuchCanIshrinkOrGrow,
  saveFinalState,
  settleAfterUpdate,
  updateContent,
  whatIsMyFlexGrow,
}

defineExpose<ExposedMembers>(exposedMembers)

function isObjectMiddleEllipsis(
  obj: ComponentPublicInstance | MiddleEllipsis,
): MiddleEllipsis | undefined {
  for (const exposedMEsymbol in exposedMembers) {
    if (!(exposedMEsymbol in obj)) {
      return undefined
    }
  }
  return obj as MiddleEllipsis
}

watch(
  slot,
  () => {
    // reacts to changes of components in our slot, and unfortunately also to changes in their props (Vue bug or feature)
    invalidateWidthCache()
    invalidateChildrenIdentities()
    innerElements.slotNonce++
  },
  { flush: 'pre' },
)
watch(
  slot,
  () => {
    // reacts to changes of components in our slot after they are mounted, and unfortunately this happens also after changes in their props
    logStep('event', 'new slot instanciated')
    invalidateChildrenIdentities()
    identifyChildren()
    nextTick(() => updateContent(0, false)) // waiting for the next tick ensures that the children are in the DOM when we start the update cycle (this slot-watcher ensured they were instantiated but not inserted in the real DOM)
  },
  { flush: 'post' },
)

watch(
  () => props.class,
  (newClassList) => {
    // reacts to changes in our list of classes
    if (newClassList === classPropsDuringLastUpdate) {
      // our watcher lags (we already updated with the correct class list)
      return
    }
    logStep('event', 'new class list received')
    invalidateTextWidthCalculationCache() // the font might have changed
    invalidateWidthCache()
    if (!amIinsideAparent.value) {
      updateContent(0, false)
    }
    else {
      logStep('signal', 'notifying my parent')
      props.meCallbackToInformParentAboutChanges!(
        SignalDirection.ChildrenToParent,
      )
    }
  },
)

watch(
  () => props.text,
  (newText) => {
    // reacts to changes of text
    if (newText === textPropsDuringLastUpdate) {
      // our watcher lags (we already updated with the correct text)
      return
    }
    logStep('event', 'new text received')
    if (!amIinsideAparent.value) {
      updateContent(0, false)
    }
    else {
      logStep('signal', 'notifying my parent')
      props.meCallbackToInformParentAboutChanges!(
        SignalDirection.ChildrenToParent,
      )
    }
  },
)

watch(
  () => props.initialFlexGrow,
  (newIFG) => {
    // reacts to changes of props initialFlexGrow
    if (newIFG === initialFlexGrowDuringLastUpdate) {
      // our watcher lags (we already updated with the correct initial flex-grow)
      return
    }
    logStep('event', 'new initial flex-grow received')
    logStep('signal', 'notifying my parent')
    props.meCallbackToInformParentAboutChanges!(
      SignalDirection.ChildrenToParent,
    )
  },
)

watch(
  () => props.ellipses,
  (newEllipses) => {
    // reacts to changes regarding the number of ellipses to use
    if (newEllipses === ellipsesPropsDuringLastUpdate) {
      // our watcher lags (we already updated with the correct value)
      return
    }
    logStep('event', 'new (array of) number(s) regarding ellipses received')
    if (amIofDefinedWidth.value) {
      // the clipping adapts the text to our width, not the other way around, so our width did not change, so we can update by ourselves (if we have a parent, a notification is useless and our siblings would spend resources updating for nothing)
      updateContent(0, false)
    }
    else {
      // our width is not defined so we have a parent
      logStep('signal', 'notifying my parent')
      props.meCallbackToInformParentAboutChanges!(
        SignalDirection.ChildrenToParent,
      )
    }
  },
)

watch(
  () => props.widthMediaqueryThreshold,
  (threshold, previousThreshold) => {
    /*  This is a workaround for a bug in Chrome (at least in April 2024).
    Here is the problem:
    When the user resizes their window and a `@media` query in the CSS changes suddenly the size of a component having a relative width
    (examples: flex-grow, width in %, auto or fr in a grid-template-columns , ...) then Chrome resizes the component in two steps.
    The first resizing is approximate for some reason and triggers the resizeObserver.
    The second resizing is definitive and accurate but does not trigger the resizeObserver, so MiddleEllipsis stays with a wrong clipping.
  */
    if (isServer || !navigator.userAgent.includes('Chrom')) {
      return
    }
    if (mediaqueryWidthListener) {
      mediaqueryWidthListener.onchange = null
    }
    if (amIinsideAparent.value || !threshold) {
      return
    }
    mediaqueryWidthListener = window.matchMedia(
      '(max-width: ' + threshold + 'px)',
    )
    mediaqueryWidthListener.onchange = catchResizingCausedByMediaquery
    if (previousThreshold) {
      // the new threshold might have passed through the current window width
      catchResizingCausedByMediaquery()
    }
  },
  { immediate: true },
)

// this function is a workaround for a bug in Chrome (see the watcher of `props.widthMediaqueryThreshold` for explanations)
function catchResizingCausedByMediaquery() {
  logStep(
    'event',
    'props.widthMediaqueryThreshold reached:',
    props.widthMediaqueryThreshold,
  )
  if (!delayedForcedUpdateIncoming) {
    delayedForcedUpdateIncoming = true
    setTimeout(() => {
      delayedForcedUpdateIncoming = false
      handleResizingEvent(true)
    }, 50)
  }
}

let resizingObserver: ResizeObserver
if (!isServer) {
  resizingObserver = new ResizeObserver(() => {
    // will react to changes of width
    if (!didTheResizingObserverFireSinceMount) {
      // we do this test because the resizing observer fires when is starts to watch, although no resizing occured at that moment
      didTheResizingObserverFireSinceMount = true
      return
    }
    logStep('event', 'resizing observer running')
    invalidateTextWidthCalculationCache() // the font might have changed, for example because the mode has been switched between mobile and desktop
    handleResizingEvent(false)
  })
}

function handleResizingEvent(force: boolean) {
  invalidateWidthCache()
  if (!amIinsideAparent.value) {
    updateContent(0, force)
  }
  else {
    const reason = determineReason(false)
    if (reason) {
      // if our resize observer lags (old resize-observer signal, we have been updated just now), we abort
      logStep('signal', 'notifying my parent for reason #', reason)
      props.meCallbackToInformParentAboutChanges!(
        SignalDirection.ChildrenToParent,
      )
    }
    else {
      logStep('good', 'parent not called because no width difference')
    }
  }
}

onMounted(() => {
  amImounted = true
  if (whatIam.value === WhatIcanBe.Error) {
    return
  }
  logStep('event', 'mounted, content given:', whatIsMyGivenContent())
  frameStyle = getComputedStyle(frameSpan.value)
  identifyChildren()
  if (doIobserveMyResizing.value) {
    didTheResizingObserverFireSinceMount = false
    resizingObserver.observe(frameSpan.value)
  }
  if (!amIinsideAparent.value) {
    updateContent(0, false)
  } // if we are inside a parent, our parent will update us because he gets mounted too
})

onBeforeUnmount(() => {
  logStep(
    'attention',
    'unmounting.',
    whatIam.value !== WhatIcanBe.Parent
      ? 'The algorithm iterated '
      + totalIterationsWhenClipping / numberOfClippings
      + ' times on average.'
      : '',
  )
  // Tests showed that watchers can be triggered by the unmounting cycle. We prevent useless recalculation to improve smoothness of the UI.
  amImounted = false
  resizingObserver.disconnect()
  if (mediaqueryWidthListener) {
    mediaqueryWidthListener.onchange = null
  }
  delayedForcedUpdateIncoming = false
})

onUnmounted(() => {
  amIreadyForUpdate = false
  // In case the font, width or slot is different at the next mount:
  invalidateTextWidthCalculationCache()
  invalidateWidthCache()
  invalidateChildrenIdentities()
})

function didMyGivenContentChange(): boolean {
  if (whatIam.value === WhatIcanBe.Parent) {
    if (innerElements.slotNonce !== lastSlotNonceWhenChecked) {
      lastSlotNonceWhenChecked = innerElements.slotNonce
      return true
    }
    return false
  }
  return props.text !== textPropsDuringLastUpdate
}

function identifyChildren(): boolean {
  if (
    innerElements.allInstanciatedElements.value.length !== slot.value.length
  ) {
    logStep('attention', 'could not identify children')
    // some children are not instanciated yet
    innerElements.isAnUpdateOrdered = true
    return false
  }
  if (innerElements.isAnUpdateOrdered) {
    // the following lines refresh our information about the inner elements passed to the slot
    innerElements.widthDefinedChildren.length = 0
    innerElements.widthUndefinedChildren.length = 0
    for (const element of innerElements.allInstanciatedElements.value) {
      const meElement = isObjectMiddleEllipsis(element)
      // if it is MiddleEllipsis, we inventor it
      if (meElement) {
        if (meElement.amIofDefinedWidth) {
          innerElements.widthDefinedChildren.push(meElement)
        }
        else {
          innerElements.widthUndefinedChildren.push(meElement)
        }
      }
    }
  }
  innerElements.isAnUpdateOrdered = false
  return true
}

function invalidateChildrenIdentities() {
  innerElements.isAnUpdateOrdered = true
}

function updateContent(additionalWidthAvailable: number, force: boolean) {
  if (whatIam.value === WhatIcanBe.Error || !amImounted || !frameSpan.value) {
    logStep(
      'attention',
      'update is impossible. amImounted and frameSpan are',
      amImounted,
      !!frameSpan.value,
    )
    return
  }
  if (whatIam.value === WhatIcanBe.Parent) {
    enterUpdateCycleAsAparent(SignalDirection.ParentToChildren, force)
  }
  else {
    currentAdditionalWidthAvailable = additionalWidthAvailable
    enterUpdateCycleAsTextClipper(force)
  }
}

function enterUpdateCycleAsAparent(
  direction: SignalDirection,
  force: boolean = false,
) {
  if (
    amIinsideAparent.value
    && direction === SignalDirection.ChildrenToParent
  ) {
    // we are a parent inside a parent, called by a child
    if (!amIreadyForUpdate) {
      logStep('signal', 'notifying my parent')
      // propagating up the refresh signal in the tree of MiddleEllipsis components
      props.meCallbackToInformParentAboutChanges!(direction)
    }
    return
  }
  if (!amImounted) {
    logStep('neutral', 'aborting update cycle: not mounted')
    // A child calls us but we are not mounted yet. No problem, we update our children after we are mounted anyway.
    return
  }
  getReadyForUpdate()
  logStep('signal', 'asking children to update and settle')
  // first we allow children with an undefined width to update their content
  for (const child of innerElements.widthUndefinedChildren) {
    child.updateContent(0, force)
  }
  // each of these children collpases their frame now to touch their text
  for (const child of innerElements.widthUndefinedChildren) {
    child.settleAfterUpdate()
  }
  // now that the undefined-width children got a width, we will allow the others to use the remaining room
  let isAchildUnclipped = false
  for (const child of innerElements.widthDefinedChildren) {
    if (child.howMuchCanIshrinkOrGrow(false) < 0) {
      isAchildUnclipped = true
      break
    }
  }
  if (!isAchildUnclipped || frameStyle.flexDirection.includes('column')) {
    for (const child of innerElements.widthDefinedChildren) {
      child.updateContent(0, force)
    }
  }
  else {
    /* The following lines handle a special case: several children have a width defined with `flex-grow`, among which at least 1 has a non-clipped text (its text is small enough to fit entirely).
    Without the following lines, after the texts are written, the flex rules would distribute the room in the span of the non-clipped text(s) to the spans of the longer text(s), after they all are written.
    That would create a gap around the clipped text(s), thus making them clipped short although there is room for more.
    The following lines detect this case and distribute the room to the children before clipping and writing, so they can clip their text longer. */
    const canUseMoreRoom: {
      child: MiddleEllipsis,
      flexGrow: number,
      growth: number,
    }[] = []
    const hasEnoughRoom: MiddleEllipsis[] = []
    let totalAdditionalRoom = 0
    let totalFlexGrow = 0
    // first, we separate children having enough room (no clipping) and those who could use this room left by the first group
    for (const child of innerElements.widthDefinedChildren) {
      const growth = child.howMuchCanIshrinkOrGrow(true)
      if (growth > 0) {
        const flexGrow = child.whatIsMyFlexGrow()
        totalFlexGrow += flexGrow
        canUseMoreRoom.push({
          child,
          flexGrow,
          growth,
        }) // For now, field `growth` represents the maximal growth of the child (due to a max-width constraint) or possibly what would allow its text not to get clipped. We will overwrite this value when we distribute the total additional room later.
      }
      else {
        totalAdditionalRoom -= growth
        hasEnoughRoom.push(child)
      }
    }
    // thanks to this sorting, the first positions hold the children that will receive more additional room than they can accept (due to max-width constraints)
    canUseMoreRoom.sort(
      (a, b) => a.growth * b.flexGrow - b.growth * a.flexGrow,
    )
    /* Note to the maintainer: a bug cannot have roots here, this sorting is proven to ensure that any child receiving too much room during the distribution sequence (see next step) is served before the others at each iteration, so that its excess can be redistributed to the next ones:
       At any iteration of the distribution sequence, a child x would be distributed too much room if and only if x.maxRoom/x.roomDistributable < 1. So, given two children a and b, it is sufficient to serve a before b if a.maxRoom/a.roomDistributable < b.maxRoom/b.roomDistributable.
       At any iteration i, x.roomDistributable is totalAdditionalRoom(i) * x.flexGrow / totalFlexGrow(i). Noticing that totalAdditionalRoom and totalFlexGrow appear on both sides of the comparison, removing them would not change the order, so the comparison can be simplified into
       a.maxRoom/a.flexGrow < b.maxRoom/b.flexGrow. Finally, as multiplications use less computing resources, the flex grow values are swapped, hence the sorting criteria above this comment block.
     */
    // Now we distribute the room available. After this step, each `cumr.growth` contains the additional room given to the child (until now, it contained its max room).
    for (const cumr of canUseMoreRoom) {
      const roomDistributable
        = (totalAdditionalRoom * cumr.flexGrow) / totalFlexGrow
      if (roomDistributable < cumr.growth) {
        cumr.growth = roomDistributable
      }
      totalAdditionalRoom -= cumr.growth
      totalFlexGrow -= cumr.flexGrow
    }
    // now the children can update with their respective rooms
    for (const cumr of canUseMoreRoom) {
      cumr.child.updateContent(cumr.growth, force)
    }
    for (const her of hasEnoughRoom) {
      her.updateContent(0, force)
    }
  }
  // now that all children adapted their text to their width, we can fill them
  for (const child of innerElements.widthDefinedChildren) {
    child.settleAfterUpdate()
  }
  // all children have been clipped and filled with their text, so they their widths are definitive, they can store their final state for future comparisons
  for (const child of innerElements.widthUndefinedChildren) {
    child.saveFinalState()
  }
  for (const child of innerElements.widthDefinedChildren) {
    child.saveFinalState()
  }
  if (!amIinsideAparent.value) {
    settleAfterUpdate()
    saveFinalState()
  }
  logStep('neutral', 'update cycle completed')
}

function enterUpdateCycleAsTextClipper(force: boolean) {
  currentText = textAfterLastUpdate.text
  getReadyForUpdate()
  if (determineReason(true) || force) {
    currentText = searchForIdealLength(
      props.text,
      getFrameWidth()
      + currentAdditionalWidthAvailable
      - ResizeObserverLagMargin,
    )
    logStep(
      'completed',
      'text clipped (with '
      + canvasContextToCalculateTextWidths.font
      + '), length difference: ',
      String(currentText.length - textAfterLastUpdate.text.length),
    )
  }
  else {
    logStep('good', 'text restored, no reclipping needed')
  }
  if (!amIinsideAparent.value) {
    settleAfterUpdate()
    saveFinalState()
  }
}

function isMyContentClipped(): boolean {
  // TODO: add a condition checking whether none of the children is clipped. Currently not required.
  return currentText !== props.text
}

/**
 * @returns a reference to an object containing the text and
 */
function calculateTextWidth(text: string | undefined): TextProperties {
  if (!text) {
    return {
      text: '',
      width: 0,
    }
  }
  if (!lastTextWidthCalculation.text) {
    // hopefully we reach this point rarely because `getComputedStyle().something` triggers a reflow (slow)
    canvasContextToCalculateTextWidths.font = frameStyle.font
  }
  if (text !== lastTextWidthCalculation.text) {
    // speed optimization (because measureText() is slow)
    lastTextWidthCalculation.text = text
    lastTextWidthCalculation.width
      = canvasContextToCalculateTextWidths.measureText(text).width
  }
  return lastTextWidthCalculation
}

function invalidateTextWidthCalculationCache() {
  lastTextWidthCalculation.text = ''
  lastTextWidthCalculation.width = 0
}

function setFrameText(text: string) {
  if (frameSpan.value) {
    frameText = text
    frameSpan.value.textContent = text
  }
}

function getFrameWidth(): number {
  if (!frameSpan.value) {
    return 0
  }
  if (!lastMeasuredFrameWidth) {
    // We do not want to read `.clientWidth` if it is unnecessary because it triggers a reflow.
    updateWidthCache(frameSpan.value.clientWidth)
  }
  return lastMeasuredFrameWidth
}

function updateWidthCache(newWidth: number) {
  lastMeasuredFrameWidth = newWidth
}

function invalidateWidthCache() {
  lastMeasuredFrameWidth = 0
}

// Use this function as rarely as possible because it triggers a reflow when it calls removeProperty() or setProperty() or reads getComputedStyle().something or reads .clientWidth
function setFrameWidth(mode: WidthMode, x: number = 0) {
  if (!frameSpan.value) {
    return
  }
  invalidateWidthCache() // We invalidate the width-cache due to the upcoming change of size, to make sure that getFrameWidth() refreshes its data
  switch (mode) {
    case WidthMode.NoFlexGrow:
      frameSpan.value.style.removeProperty('flex-grow')
      break
    case WidthMode.NoWidth: // not used currently but this makes the function ready for future needs / modes
      frameSpan.value.style.removeProperty('width')
      break
    case WidthMode.FixedFlexGrow:
      frameSpan.value.style.setProperty('flex-grow', String(x))
      break
    case WidthMode.FixedWidth: {
      // not used currently but this makes the function ready for future needs / modes
      const minWidth = parseFloat(frameStyle.minWidth)
      const maxWidth = parseFloat(frameStyle.maxWidth)
      if (x < minWidth) {
        x = minWidth
      }
      if (x > maxWidth) {
        x = maxWidth
      }
      updateWidthCache(x)
      frameSpan.value.style.setProperty('width', String(x) + 'px')
    }
  }
}

function determineReason(
  considerThatTheChangeAffectsMeOnly: boolean,
): UpdateReason {
  let reason: UpdateReason
  const gaps = calculateGaps()
  if (
    gaps.before === undefined
    || didMyGivenContentChange()
    || whatIam.value === WhatIcanBe.Parent
    || gaps.now < 0
    || gaps.before < 0
  ) {
    reason = UpdateReason.GapChangeToBeDetermined
  }
  else {
    let changeMightNotRequireAnUpdate = gaps.now >= ResizeObserverLagMargin // the current content still fits the frame
    if (gaps.now < gaps.before) {
      reason = UpdateReason.GapChangeMinus
    }
    else if (gaps.now > gaps.before) {
      reason = UpdateReason.GapChangePlus
      changeMightNotRequireAnUpdate &&= !isMyContentClipped() // despite the wider gap, the content will not be clipped longer because it is already not clipped
    }
    else {
      reason = UpdateReason.None
    }
    if (considerThatTheChangeAffectsMeOnly && changeMightNotRequireAnUpdate) {
      reason = UpdateReason.None
    }
  }
  logStep(
    'neutral',
    [
      'my gap is fine as it is.',
      'my gap changed (to be determined).',
      'my gap decreased.',
      'my gap increased.',
    ][reason],
    'Gaps:',
    gaps,
  )
  return reason

  function calculateGaps(): { before: number | undefined,
    now: number, } {
    // TODO: If needed, calculate the actual gaps when we are a parent (frame width - sum of child widths). Currently not required.
    let before: number | undefined
    const frameWhidthToCompareTo = amIreadyForUpdate
      ? widthAvailableDuringLastUpdate
      : frameWidthAfterLastUpdate
    const now
      = getFrameWidth()
      + (amIreadyForUpdate ? currentAdditionalWidthAvailable : 0)
      - calculateTextWidth(currentText).width
    if (frameWhidthToCompareTo) {
      before = frameWhidthToCompareTo - textAfterLastUpdate.width
    }
    else {
      before = undefined
    }
    return {
      before,
      now,
    }
  }
}

function whatIsMyFlexGrow(): number {
  return Number(frameStyle.flexGrow) || 0
}

/**
 * Assuming that the content is not clipped, this tells how much the frame could shrink or grow if it had to be as large as the content (or hit min-width or max-width).
 * @returns If `accurate` is `true`: positive means I can grow so much, negative means I can shrink so much. If `accurate` is `false`: positive means the content will be clipped, negative means the content fits entirely.
 */
function howMuchCanIshrinkOrGrow(accurate: boolean): number {
  const widthRightNow = getFrameWidth()
  let withoutRestriction: number
  if (whatIam.value === WhatIcanBe.Parent) {
    withoutRestriction = 0
    for (const child of innerElements.widthDefinedChildren) {
      withoutRestriction += child.howMuchCanIshrinkOrGrow(true)
    }
  }
  else {
    withoutRestriction
      = calculateTextWidth(props.text).width
      - (widthRightNow - ResizeObserverLagMargin)
  }
  if (!accurate) {
    return withoutRestriction
  }
  if (frameStyle.flexGrow === '') {
    return 0
  }
  if (withoutRestriction >= 0) {
    const limit = parseFloat(frameStyle.maxWidth) || Number.MAX_SAFE_INTEGER
    return widthRightNow + withoutRestriction <= limit
      ? withoutRestriction
      : limit - widthRightNow
  }
  else {
    const limit = parseFloat(frameStyle.minWidth) || 0
    return widthRightNow + withoutRestriction >= limit
      ? withoutRestriction
      : limit - widthRightNow
  }
}

function getReadyForUpdate() {
  if (amIreadyForUpdate) {
    // we have a parent and he already called this function
    return
  }
  amIreadyForUpdate = true
  if (!amIofDefinedWidth.value) {
    // our undefined width requires us to get a width before clipping the text. Note that settleAfterUpdate() will undefine our width later.
    setFrameWidth(WidthMode.FixedFlexGrow, props.initialFlexGrow)
  }
  if (whatIam.value !== WhatIcanBe.Parent) {
    logStep('neutral', 'getting ready for update')
    setFrameText('') // better done after setFrameWidth() for performance reasons
  }
  else {
    identifyChildren()
    logStep('signal', 'asking children to get ready')
    for (const child of innerElements.widthDefinedChildren) {
      // All children of defined width must be prepared first. Preparing the undefined-width children first would change the width of the defined-width ones (because of the initial flex-grows of the undefined-width ones), thus making their reasons unreliable.
      child.getReadyForUpdate()
    }
    for (const child of innerElements.widthUndefinedChildren) {
      // Children having an undefined will find now an initial (and temporary) width
      child.getReadyForUpdate()
    }
  }
  invalidateWidthCache()
}

function settleAfterUpdate() {
  classPropsDuringLastUpdate = props.class || ''
  textPropsDuringLastUpdate = props.text || ''
  initialFlexGrowDuringLastUpdate = props.initialFlexGrow
  ellipsesPropsDuringLastUpdate = props.ellipses
  widthAvailableDuringLastUpdate
    = getFrameWidth() + currentAdditionalWidthAvailable
  if (whatIam.value !== WhatIcanBe.Parent) {
    setFrameText(currentText)
  }
  if (!amIofDefinedWidth.value) {
    // our undefined width required us to get a width before clipping the text. Now we must undefine our width.
    setFrameWidth(WidthMode.NoFlexGrow)
  }
  amIreadyForUpdate = false
  logStep('neutral', 'settled')
}

function saveFinalState() {
  invalidateWidthCache()
  frameWidthAfterLastUpdate = getFrameWidth()
  textAfterLastUpdate = { ...calculateTextWidth(currentText) }
}

function logStep(
  color: 'attention' | 'completed' | 'event' | 'good' | 'neutral' | 'signal',
  msg: string,
  ...others: any[]
) {
  if (!DEBUG) {
    return
  }
  const parentInParent
    = whatIam.value === WhatIcanBe.Parent && amIinsideAparent.value
  let common = ''

  if (whatIam.value === WhatIcanBe.Standalone) {
    common += '\u001B[90m'
  }
  common
    += whatIam.value === WhatIcanBe.Child ? '    ' : parentInParent ? '  ' : ''
  common += [
    'Error',
    'Parent',
    'Child',
    'Standalone',
  ][whatIam.value]
  if (whatIam.value !== WhatIcanBe.Parent) {
    common += ' "' + (props.text as string).slice(0, 8) + '…"'
  }
  common
    += (amIofDefinedWidth.value ? ' (defined' : ' (undef')
    + ' width cached: '
    + lastMeasuredFrameWidth
    + ') '
  switch (color) {
    case 'attention':
      msg = '\u001B[31m' + msg
      break
    case 'completed':
      msg = '\u001B[35m' + msg
      break
    case 'event':
      msg = '\u001B[33m' + msg
      break
    case 'good':
      msg = '\u001B[34m' + msg
      break
    case 'signal':
      msg = '\u001B[32m' + msg
      break
    default:
      msg = '\u001B[0m' + msg
  }
  const writer = console
  writer.log(common + msg, ...others)
}

function whatIsMyGivenContent(): any {
  return whatIam.value !== WhatIcanBe.Parent
    ? props.text
      ? 'text'
      : 'no text'
    : innerElements.allInstanciatedElements.value
}

function searchForIdealLength(
  originalText: string = '',
  targetWidth: number,
): string {
  let current = calculateTextWidth(originalText)
  // we search for the longest clipped text which fits in the target width
  if (current.width > targetWidth) {
    let maxWidth = current.width
    let maxLength = current.text.length
    let minWidth = 0
    let minLength = 0
    while (minLength < maxLength - 1) {
      let averageCharWidthBetweenCurrentAndBound: number
      if (current.width > targetWidth) {
        maxLength = current.text.length
        maxWidth = current.width
        averageCharWidthBetweenCurrentAndBound
          = (current.width - minWidth) / (current.text.length - minLength)
      }
      else {
        minLength = current.text.length
        minWidth = current.width
        averageCharWidthBetweenCurrentAndBound
          = (maxWidth - current.width) / (maxLength - current.text.length)
      }
      // The following block writes in `lengthToTry` an estimation of the length that the clipped text should have to reach our target width.
      // This estimation speeds up considerably the dichotomic search by guiding it towards the optimal length. The way we use it preserves optimality.
      let estimatedLengthExcess
        = (current.width - targetWidth) / averageCharWidthBetweenCurrentAndBound
      if (estimatedLengthExcess > 0 && estimatedLengthExcess <= 0.5) {
        estimatedLengthExcess += 0.5
      } // this avoids slight overflows, due to the `Math.round` just below
      estimatedLengthExcess = Math.round(estimatedLengthExcess)
      let lengthToTry = current.text.length - estimatedLengthExcess
      // If the estimation exceeds the range of possible widths, we default to a classical dichotomic search for the current iteration:
      if (lengthToTry < minLength || lengthToTry >= maxLength) {
        // ...this is why optimality is preserved.
        lengthToTry = Math.floor((minLength + maxLength) / 2)
      }
      else if (estimatedLengthExcess === 0) {
        break
      }
      current = calculateTextWidth(
        clipText(
          originalText,
          lengthToTry,
          numberOfEllipsesSetInProps(lengthToTry),
        ),
      )
      totalIterationsWhenClipping++
    }
    numberOfClippings++
  }
  return current.text
}

function clipText(
  originalText: string,
  room: number,
  nEllipses: number,
): string {
  if (room <= 0 || isNaN(room) || !originalText) {
    return ''
  }
  if (originalText.length <= room) {
    return originalText
  }
  if (room === 1) {
    return '…'
  }
  if (nEllipses <= 0) {
    nEllipses = 1
  }
  if (nEllipses > originalText.length - room) {
    // Each ellipsis must represent at least one clipped character, otherwise it will just hide a displayable character.
    nEllipses = originalText.length - room
  }
  if (nEllipses > room - nEllipses) {
    // There are more ellipses than visible characters (for example 5 chars including 3 ellipses would give A……D…)
    nEllipses = Math.floor(room / 2) // thus we guarantee at least one character per ellipsis (giving AB…E… with our example) to avoid useless a loss of information for the user (and ugly doubled ellipses)
  }

  // simple and fast
  if (nEllipses === 1) {
    const midL = Math.floor(room / 2)
    const r = 1 - (room % 2)
    return (
      originalText.slice(0, midL)
      + '…'
      + originalText.slice(originalText.length - midL + r, originalText.length)
    )
  }
  // complicated and slower
  const nBlocks = nEllipses + 1
  // First, we extract `nBlocks` blocks from the original text. The extraction algorithm guarantees that the sum of their lengths is `room`. Their lengths vary for arithmetic reasons but are as similar as possible.
  type Block = { start: number,
    visibleLength: number, }
  const blocks: Block[] = []
  let totalToExtract = room
  let totalToSkip = originalText.length - room
  let skipNow = 0
  let s = nEllipses
  let o = 0
  for (let b = nBlocks; b > 0; b--) {
    const extractedNow = Math.round(totalToExtract / b) // do not floor here
    totalToExtract -= extractedNow
    o += skipNow
    blocks.push({
      start: o,
      visibleLength: extractedNow,
    })
    o += extractedNow
    skipNow = Math.floor(totalToSkip / s) // do not round here
    totalToSkip -= skipNow
    s--
  }
  // Now, we decide whether the first and/or last letter of each block should be replaced with an ellipsis. We try to create symmetry (to get something like AB…E…HI rather than AB…EF…I) because the two ends of a text are easier to catch and compare for a human eye (and we are "MiddleEllipsis" so the ellipses must look centered)
  enum Side {
    EndOfB,
    BeginningBp1,
  }
  const sides: Side[] = [] // each element of index b indicates the side of the ellipsis between blocks b and b+1
  const symmAxis = Math.floor(nBlocks / 2)
  // first, we decide it over the first half of blocks
  let bLeft = 0
  let bRight = nBlocks - 1
  while (bLeft < symmAxis) {
    if (blocks[bLeft].visibleLength > blocks[bRight].visibleLength) {
      // we have a look at the opposite bloc, to try to create symmetry
      sides.push(Side.EndOfB)
      blocks[bLeft].visibleLength--
    }
    else {
      sides.push(Side.BeginningBp1)
      blocks[bLeft + 1].visibleLength--
    }
    bLeft++
    bRight--
  }
  // now, we decide it over the second half of blocks
  bRight = symmAxis + 1
  bLeft = nBlocks - 1 - bRight
  while (bRight < nBlocks) {
    if (
      blocks[bRight].visibleLength > blocks[bLeft].visibleLength
      || blocks[bRight - 1].visibleLength <= 1
    ) {
      // we have a look at the opposite bloc, trying to create symmetry with the first half while forbidding two ellipses in blocks of size 2
      sides.push(Side.BeginningBp1)
      blocks[bRight].visibleLength--
    }
    else {
      sides.push(Side.EndOfB)
      blocks[bRight - 1].visibleLength--
    }
    bRight++
    bLeft--
  }
  // Finally we create the result with the information that we calculated
  let result = originalText.slice(blocks[0].start, blocks[0].visibleLength)
  for (let b = 1; b < nBlocks; b++) {
    const start = blocks[b].start + (sides[b - 1] === Side.EndOfB ? 0 : 1)
    result += '…' + originalText.slice(start, start + blocks[b].visibleLength)
  }
  return result
}

function numberOfEllipsesSetInProps(textLength: number): number {
  if (props.ellipses === undefined) {
    return 1
  }
  if (typeof props.ellipses === 'number') {
    return props.ellipses <= 0 ? 1 : props.ellipses
  }
  let result = 1
  for (const threshold of props.ellipses) {
    if (textLength <= threshold) {
      return result
    }
    result++
  }
  return result
}

const frameClassList = computed(
  () => 'middle-ellipsis-root-frame ' + props.class,
)
</script>

<template>
  <span
    ref="frameSpan"
    :class="frameClassList"
  >
    {{ frameText }}
    <!--
      The text above is not reactive because its only purpose is to provide a content during isServer. During CSR,
      MiddleEllipsis clips and overwrites the text
      of the frame with a direct assignment (frameSpan.value.textContent = ...), which has an immediate effect within
      one reflow unlike reactive properties.
      If for some reason Vue were to rewrite the content of the frame with the variable above while the component
      is mounted, this would not cause any problem
      because it always mirrors what has been directly assigned to the frame.
      The following line mounts our slot if we have one.
      To inform MiddleEllipsis components that they are children and give them the ability to send us information,
      we add a props. Also, we get a ref to each instantiated element.
    -->
    <component
      :is="slotElem"
      v-for="slotElem of slot"
      :key="slotElem"
      :ref="innerElements.allInstanciatedElements"
      :me-callback-to-inform-parent-about-changes="enterUpdateCycleAsAparent"
    />
  </span>
</template>

<style lang="scss">
.middle-ellipsis-root-frame {
  display: inline-flex;
  position: relative;
  box-sizing: border-box;
  vertical-align: middle;
  flex-wrap: nowrap;
  white-space: nowrap;
  overflow: hidden;
}
</style>
