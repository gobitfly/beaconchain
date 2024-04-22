<script setup lang="ts">
import { type ComponentPublicInstance, warn } from 'vue'

const DEBUG = false // Use Chromium or Chrome. Firefox will show messages with broken indentation, illegible codes and no color differenciating the types of the messages.

interface ExportedMembers {
  myInstanceId: ComputedRef<number>,
  amIofDefinedWidth: ComputedRef<boolean>,
  getReadyForUpdate: typeof getReadyForUpdate,
  updateContent: typeof updateContent,
  settleAfterUpdate: typeof settleAfterUpdate,
  saveFinalState: typeof saveFinalState,
  enterUpdateCycleAsAparent: typeof enterUpdateCycleAsAparent
}

interface MiddleEllipsis extends ComponentPublicInstance, ExportedMembers {}

enum WhatIcanBe {
  Error = 0,
  Parent,
  Child,
  Standalone
}

enum WidthMode {
  NoFlexGrow,
  NoWidth,
  FixedFlexGrow,
  FixedWidth
}

// This identifiers will tell us how the gap between the end of the text and the right edge of the frame changes
enum UpdateReason {
  None = 0,
  GapChangeToBeDetermined,
  GapChangeMinus,
  GapChangePlus
}

type TextProperties = { text: string, width: number }

const ResizeObserverLagMargin = 1.5 // This safety margin is important, because the resizing observer happens to lag. If a small decrease of width making the frame as large as its content does not trigger the observer, then it will not fire anymore because the frame cannot shrink anymore.

const props = defineProps<{
  text?: string,
  initialFlexGrow?: number, // if the component has no defined size (meaning that its width collapses to 0 when it contains nothing) then you must set a value in this props
  ellipses? : number | number[], // If number: number of ellipses to use (the same for any room available), 1 by default. If array, its meaning is: [room above which two `…` are used, room above which three `…` are used, and so on]. Ex: [8,30,100] tells the component to use one ellipsis if there is room for 8 characters or less, or two ellipses between 9 and 30 characters, and so on
  meCallbackToInformParentAboutChanges?: typeof enterUpdateCycleAsAparent, // for internal use, to inform this instance that it belongs to a parent MiddleEllipsis component
  meInstanceId?: number
  class? : string // hack to make the list of classes reactive
}>()

const _s = useSlots() // Not meant to be used directly. Use the reactive variable `slot` defined just below:
const slot = computed(() => _s.default ? _s.default() : []) // `slot`s is always an array, empty if there is no slot

const innerElements = {
  allInstanciatedElements: ref<(ComponentPublicInstance | MiddleEllipsis)[]>([]), // Instanciated elements from our slot. This array is filled by Vue in the <template>.
  // The following arrays are filled by us, each time the slot is modified:
  widthDefinedChildren: [] as MiddleEllipsis[], // List of instanciated elements from our slot that are MiddleEllipsis children with a defined width.
  widthUndefinedChildren: [] as MiddleEllipsis[], // List of instanciated elements from our slot that are MiddleEllipsis children with an undefined width.
  isAnUpdateOrdered: true,
  slotNonce: 0
}
const frameSpan = ref<HTMLSpanElement>(null as unknown as HTMLSpanElement)

let classPropsDuringLastUpdate = props.class || ''
let textPropsDuringLastUpdate = props.text || ''
let initialFlexGrowDuringLastUpdate = 0
let textAfterLastUpdate : TextProperties = { text: '', width: 0 }
let frameWidthDuringLastUpdate = 0 // used by determineReason() to find out why an update is needed, during the update process
let frameWidthAfterLastUpdate = 0 // used by determineReason() to find out why an update is needed, outside the update process
let lastMeasuredFrameWidth = 0
let currentText = ''
const canvasContextToCalculateTextWidths = document.createElement('canvas').getContext('2d') as CanvasRenderingContext2D
const lastTextWidthCalculation: TextProperties = { text: '', width: 0 }
let amImounted = false
let didTheResizingObserverFireSinceMount = false
let amIreadyForUpdate = false // 1. Our parent can call function getReadyForUpdate() as we can, so we use this variable to prevent multiple executions of it in a row. 2. It is also useful in our resizing observer to know whether our parent started our update process so we do not inform it about our changes if our observer fires late.
let lastSlotNonceWhenChecked = -1

let numberOfClippings = 0
let totalIterationsWhenClipping = 0

const myInstanceId = computed(() => {
  return props.meInstanceId === undefined ? -1 : props.meInstanceId
})

const amIofDefinedWidth = computed(() => {
  // TODO: Maybe check whether the width is defined in the CSS of the component if-and-only-if props.initialFlexGrow is not set.
  //       Problem if done: it would be a slow operation at execution time just to provide a security against the programmer during development (because during execution it causes bugs anyway).
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
    warn('MiddleEllipsis cannot have an undefined width because it is outside a MiddleEllipsis container.')
    return WhatIcanBe.Error
  }
  if (slot.value[0]) {
    if (props.text) {
      warn('MiddleEllipsis cannot receive any content in props `text` because it is a container.')
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

const exportedMembers : ExportedMembers = {
  myInstanceId,
  amIofDefinedWidth,
  getReadyForUpdate,
  updateContent,
  settleAfterUpdate,
  saveFinalState,
  enterUpdateCycleAsAparent
}

function isObjectMiddleEllipsis (obj : MiddleEllipsis | ComponentPublicInstance) : MiddleEllipsis | undefined {
  for (const exportedMEsymbol in exportedMembers) {
    if (!(exportedMEsymbol in obj)) {
      return undefined
    }
  }
  return obj as MiddleEllipsis
}

watch(slot, () => { // reacts to changes of components in our slot, and unfortunately also to changes in their props
  invalidateWidthCache()
  invalidateChildrenIdentities()
  innerElements.slotNonce++
}, {
  flush: 'pre'
})
watch(slot, () => { // reacts to changes of components in our slot after they are mounted, and unfortunately this happens also after changes in their props
  logStep('yellow', 'new slot instanciated')
  invalidateChildrenIdentities()
  identifyChildren()
  nextTick(() => enterUpdateCycleAsAparent()) // waiting for the next tick ensures that the children are in the DOM when we start the update cycle (unfortunately, this slot-watcher ensured they were instanciated but not inserted in the real DOM)
}, {
  flush: 'post'
})

watch(() => props.class, (newClassList) => { // reacts to changes in our list of classes
  if (newClassList === classPropsDuringLastUpdate) {
    // our watcher lags (we already updated with the correct class list)
    return
  }
  logStep('yellow', 'new class list received')
  invalidateTextWidthCalculationCache() // the font might have changed
  invalidateWidthCache()
  if (!amIinsideAparent.value) {
    updateContent()
  } else // No self update is allowed, we must ask our parent to update us.
    if (!amIreadyForUpdate) { // if our parent is updating us already, we abort
      logStep('green', 'notifying my parent')
      props.meCallbackToInformParentAboutChanges!(myInstanceId.value)
    } else { logStep('blue', 'parent not called') }
})

watch(() => props.text, (newText) => { // reacts to changes of text
  if (newText === textPropsDuringLastUpdate) {
    // our watcher lags (we already updated with the correct text)
    return
  }
  logStep('yellow', 'new text received')
  if (amIofDefinedWidth.value) {
    // the clipping adapts the text to our width, not the other way around, so our width did not change, so we can update by ourselves (if we have a parent, a notification is useless and our siblings would spend resources updating for nothing)
    updateContent()
  } else // Our width is not defined so we have a parent.
    if (!amIreadyForUpdate) { // if our parent is updating us already, we abort
      logStep('green', 'notifying my parent')
      props.meCallbackToInformParentAboutChanges!(myInstanceId.value)
    } else { logStep('blue', 'parent not called') }
})

watch(() => props.initialFlexGrow, (newFG) => { // reacts to changes of props initialFlexGrow
  if (newFG === initialFlexGrowDuringLastUpdate) {
    // our watcher lags (we already updated with the correct initial flex-grow)
    return
  }
  logStep('yellow', 'new initial flex-grow received')
  // No self update is allowed, we must ask our parent to update us.
  if (!amIreadyForUpdate) { // if our parent is updating us already, we abort
    logStep('green', 'notifying my parent')
    props.meCallbackToInformParentAboutChanges!(myInstanceId.value)
  } else { logStep('blue', 'parent not called') }
})

watch(() => props.ellipses, () => { // reacts to changes regarding the number of ellipses to use
  logStep('yellow', 'new (array of) number(s) regarding ellipses received')
  if (amIofDefinedWidth.value) {
    // the clipping adapts the text to our width, not the other way around, so our width did not change, so we can update by ourselves (if we have a parent, a notification is useless and our siblings would spend resources updating for nothing)
    updateContent()
  } else // Our width is not defined so we have a parent.
    if (!amIreadyForUpdate) { // if our parent is updating us already, we abort
      logStep('green', 'notifying my parent')
      props.meCallbackToInformParentAboutChanges!(myInstanceId.value)
    } else { logStep('blue', 'parent not called') }
})

const resizingObserver = new ResizeObserver(() => { // will react to changes of width
  if (!didTheResizingObserverFireSinceMount) {
    // we do this test because the resizing observer fires when is starts to watch, although no resizing occured at that moment
    didTheResizingObserverFireSinceMount = true
    return
  }
  invalidateWidthCache()
  logStep('yellow', 'resizing observer running')
  if (!amIinsideAparent.value) {
    updateContent()
  } else
    if (!amIreadyForUpdate) { // if our parent is updating us already, we abort
      const reason = determineReason(false)
      if (reason) { // if our resize observer lags (old resize-observer signal, we have been updated just now), we abort
        logStep('green', 'notifying my parent for reason #', reason)
        props.meCallbackToInformParentAboutChanges!(myInstanceId.value)
      } else { logStep('blue', 'parent not called because no width difference') }
    }
})

onMounted(() => {
  amImounted = true
  if (whatIam.value === WhatIcanBe.Error) {
    return
  }
  logStep('yellow', 'mounted, content given:', whatIsMyGivenContent())
  identifyChildren()
  if (doIobserveMyResizing.value) {
    didTheResizingObserverFireSinceMount = false
    resizingObserver.observe(frameSpan.value)
  }
  if (!amIinsideAparent.value) {
    updateContent()
  } // if we are inside a parent, our parent will update us because he gets mounted too
})

onBeforeUnmount(() => {
  logStep('red', 'unmounting.', whatIam.value !== WhatIcanBe.Parent ? 'The algorithm iterated ' + totalIterationsWhenClipping / numberOfClippings + ' times on average.' : '')
  // Tests showed that watchers can be triggered by the unmounting cycle. We prevent useless recalculation to improve smoothness of the UI.
  amImounted = false
  resizingObserver.disconnect()
})

onUnmounted(() => {
  amIreadyForUpdate = false
  // In case the font, width or slot is different at the next mount:
  invalidateTextWidthCalculationCache()
  invalidateWidthCache()
  invalidateChildrenIdentities()
})

function didMyGivenContentChange () : boolean {
  if (whatIam.value === WhatIcanBe.Parent) {
    if (innerElements.slotNonce !== lastSlotNonceWhenChecked) {
      lastSlotNonceWhenChecked = innerElements.slotNonce
      return true
    }
    return false
  }
  return props.text !== textPropsDuringLastUpdate
}

function areChildrenIdentified () : boolean {
  return !innerElements.isAnUpdateOrdered && innerElements.allInstanciatedElements.value.length === slot.value.length
}

function identifyChildren () : boolean {
  if (innerElements.allInstanciatedElements.value.length !== slot.value.length) {
    logStep('red', 'could not identify children')
    // some children are not instanciated yet
    innerElements.isAnUpdateOrdered = true
    return false
  }
  if (innerElements.isAnUpdateOrdered) {
  // the following lines refresh our information about the inner elements passed to the slot, and then we call updateContent() to manage their instances
    innerElements.widthDefinedChildren.length = 0
    innerElements.widthUndefinedChildren.length = 0
    for (const element of innerElements.allInstanciatedElements.value) {
      const meElement = isObjectMiddleEllipsis(element)
      // if it is MiddleEllipsis, we inventor it
      if (meElement) {
        if (meElement.amIofDefinedWidth) {
          innerElements.widthDefinedChildren.push(meElement)
        } else {
          innerElements.widthUndefinedChildren.push(meElement)
        }
      }
    }
  }
  innerElements.isAnUpdateOrdered = false
  return true
}

function invalidateChildrenIdentities () {
  innerElements.isAnUpdateOrdered = true
}

function updateContent () {
  if (whatIam.value === WhatIcanBe.Error || !amImounted || !frameSpan.value) {
    logStep('red', 'update is impossible. amImounted and frameSpan are', amImounted, !!frameSpan.value)
    return
  }
  if (whatIam.value === WhatIcanBe.Parent) {
    enterUpdateCycleAsAparent()
  } else {
    enterUpdateCycleAsTextclipper()
  }
}

function enterUpdateCycleAsAparent (childId? : number) {
  if (amIinsideAparent.value) {
    // if we are here, it means we are a parent inside a parent
    if (!amIreadyForUpdate) {
      logStep('green', 'notifying my parent')
      // propagating up the refresh signal in the tree of MiddleEllipsis components
      props.meCallbackToInformParentAboutChanges!(myInstanceId.value)
    }
    return
  }
  if (!amImounted) {
    logStep('normal', 'aborting update cycle: not mounted')
    // A child calls us but we are not mounted yet. No problem, we update our children after we are mounted anyway.
    return
  }
  if (!areChildrenIdentified()) {
    // we do not know all our children yet (they are beeing mounted or have been too recently)
    warn('MiddleEllipsis entered an update cycle as parent but its children are not all known yet. This is an internal bug. Child #' + childId)
    return
  }
  identifyChildren()
  getReadyForUpdate()
  logStep('green', 'asking children to update and settle')
  // first we allow children with an undefined width to update their content
  for (const child of innerElements.widthUndefinedChildren) {
    child.updateContent()
  }
  // each of these children collpases their frame now to touch their text
  for (const child of innerElements.widthUndefinedChildren) {
    child.settleAfterUpdate()
  }
  /*
  TODO: one last visual bug to fix :)
  Solution:
   implement and expose calculateGapsWithOriginalText()
   updateContent() should take a new argument (additional room)
   Before updating: if at least one defined-width child with a flex-grow value (read css) has a large gap (which means it will not clip)
     sum the gaps of those.
     spread this additional room over the width-defined children with no gap AND a flex-grow value
  */
  // now that the undefined-width children got a width, we allow the others to use the remaining room
  for (const child of innerElements.widthDefinedChildren) {
    child.updateContent()
  }
  // now that they adapted their text to their width, we can fill them, their text is decided so their will not influence each other
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
  logStep('normal', 'update cycle completed')
}

function enterUpdateCycleAsTextclipper () {
  currentText = textAfterLastUpdate.text
  getReadyForUpdate()
  if (determineReason(true)) {
    currentText = searchForIdealLength(props.text, getFrameWidth() - ResizeObserverLagMargin)
    logStep('purple', 'text clipped, length difference: ', String(currentText.length - textAfterLastUpdate.text.length))
  } else {
    logStep('blue', 'text restored, no reclipping needed')
  }
  if (!amIinsideAparent.value) {
    settleAfterUpdate()
    saveFinalState()
  }
}

function isMyContentClipped () : boolean {
  // TODO: add a condition checking whether none of the children is clipped. Currently not required.
  return currentText !== props.text
}

/**
 * @returns a reference to an object containing the text and
 */
function calculateTextWidth (text: string | undefined): TextProperties {
  if (!text) {
    return { text: '', width: 0 }
  }
  if (!lastTextWidthCalculation.text) {
    // hopefully we reach this point rarely because `getComputedStyle().something` triggers a reflow (slow)
    canvasContextToCalculateTextWidths.font = getComputedStyle(frameSpan.value).font
  }
  if (text !== lastTextWidthCalculation.text) { // speed optimization (because measureText() is slow)
    lastTextWidthCalculation.text = text
    lastTextWidthCalculation.width = canvasContextToCalculateTextWidths.measureText(text).width
  }
  return lastTextWidthCalculation
}

function invalidateTextWidthCalculationCache () {
  lastTextWidthCalculation.text = ''
  lastTextWidthCalculation.width = 0
}

function setFrameText (text: string) {
  if (frameSpan.value) {
    frameSpan.value.textContent = text
  }
}

function getFrameWidth (): number {
  if (!frameSpan.value) {
    return 0
  }
  if (!lastMeasuredFrameWidth) { // We do not want to read `.clientWidth` if it is unnecessary because it triggers a reflow.
    updateWidthCache(frameSpan.value.clientWidth)
  }
  return lastMeasuredFrameWidth
}

function updateWidthCache (newWidth : number) {
  lastMeasuredFrameWidth = newWidth
}

function invalidateWidthCache () {
  lastMeasuredFrameWidth = 0
}

// Use this function as rarely as possible because it triggers a reflow when it calls removeProperty() or setProperty() or reads getComputedStyle().something or reads .clientWidth
function setFrameWidth (mode: WidthMode, x: number = 0) {
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
    case WidthMode.FixedWidth: { // not used currently but this makes the function ready for future needs / modes
      const minWidth = parseFloat(getComputedStyle(frameSpan.value).minWidth)
      const maxWidth = parseFloat(getComputedStyle(frameSpan.value).maxWidth)
      if (x < minWidth) { x = minWidth }
      if (x > maxWidth) { x = maxWidth }
      updateWidthCache(x)
      frameSpan.value.style.setProperty('width', String(x) + 'px')
    }
  }
}

function determineReason (considerThatTheChangeAffectMeOnly : boolean) : UpdateReason {
  let reason : UpdateReason
  const gaps = calculateGaps()
  if (gaps.before === undefined || didMyGivenContentChange() || whatIam.value === WhatIcanBe.Parent) {
    reason = UpdateReason.GapChangeToBeDetermined
  } else {
    let changeMightNotRequireAnUpdate = gaps.now >= ResizeObserverLagMargin // the content still fits the frame
    if (gaps.now < gaps.before) {
      reason = UpdateReason.GapChangeMinus
    } else if (gaps.now > gaps.before) {
      reason = UpdateReason.GapChangePlus
      changeMightNotRequireAnUpdate &&= !isMyContentClipped() // despite the wider gap, the content cannot be clipped longer because it is already not clipped
    } else {
      reason = UpdateReason.None
    }
    if (considerThatTheChangeAffectMeOnly && changeMightNotRequireAnUpdate) {
      reason = UpdateReason.None
    }
  }
  logStep('normal', ['my gap is fine as it is.', 'my gap changed (to be determined).', 'my gap decreased.', 'my gap increased.'][reason], 'Gaps:', gaps)
  return reason

  function calculateGaps () : {before : number|undefined, now : number} {
    // TODO: If needed, calculate the actual gaps when we are a parent (frame width - sum of child widths). Currently not required.
    let before : number | undefined
    const frameWhidthToCompareTo = amIreadyForUpdate ? frameWidthDuringLastUpdate : frameWidthAfterLastUpdate
    const now = getFrameWidth() - calculateTextWidth(currentText).width
    if (frameWhidthToCompareTo) {
      before = frameWhidthToCompareTo - textAfterLastUpdate.width
    } else {
      before = undefined
    }
    return { before, now }
  }
}

// returns the text that was in the frame before it got emptied
function getReadyForUpdate () {
  if (amIreadyForUpdate) {
    // we have a parent and he already called this function
    return
  }
  amIreadyForUpdate = true
  if (!amIofDefinedWidth.value) {
    // our undefined width requires us to get a width before clipping the text. Note that settleAfterUpdate() will undefine our width later.
    setFrameWidth(WidthMode.FixedFlexGrow, props.initialFlexGrow)
    initialFlexGrowDuringLastUpdate = props.initialFlexGrow!
  }
  if (whatIam.value !== WhatIcanBe.Parent) {
    logStep('normal', 'getting ready for update')
    setFrameText('') // better done after setFrameWidth() for performance reasons
  } else {
    identifyChildren()
    logStep('green', 'asking children to get ready')
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

function settleAfterUpdate () {
  classPropsDuringLastUpdate = props.class || ''
  textPropsDuringLastUpdate = props.text || ''
  frameWidthDuringLastUpdate = getFrameWidth()
  if (whatIam.value !== WhatIcanBe.Parent) {
    setFrameText(currentText)
  }
  if (!amIofDefinedWidth.value) {
    // our undefined width required us to get a width before clipping the text. Now we must undefine our width.
    setFrameWidth(WidthMode.NoFlexGrow)
  }
  amIreadyForUpdate = false
  logStep('normal', 'settled')
}

function saveFinalState () {
  invalidateWidthCache()
  frameWidthAfterLastUpdate = getFrameWidth()
  textAfterLastUpdate = { ...calculateTextWidth(currentText) }
}

function logStep (color : 'normal'|'red'|'yellow'|'green'|'blue'|'purple', msg : string, a? : any, b? : any, c? : any) {
  if (DEBUG) {
    const parentInParent = whatIam.value === WhatIcanBe.Parent && amIinsideAparent.value
    let common = ''

    if (whatIam.value === WhatIcanBe.Standalone) {
      common += '\u001B[90m'
    }
    common += whatIam.value === WhatIcanBe.Child ? '    ' : (parentInParent ? '  ' : '')
    common += ['Error', 'Parent', 'Child', 'Standalone'][whatIam.value]
    if (whatIam.value === WhatIcanBe.Child || parentInParent) {
      common += ' #' + myInstanceId.value
    }
    if (whatIam.value !== WhatIcanBe.Parent) {
      common += ' "' + (props.text as string).slice(0, 8) + '…"'
    }
    common += (amIofDefinedWidth.value ? ' (defined' : ' (undef') + ' width cached: ' + lastMeasuredFrameWidth + ') '
    switch (color) {
      case 'red' : msg = '\u001B[31m' + msg; break
      case 'yellow' : msg = '\u001B[33m' + msg; break
      case 'green' : msg = '\u001B[32m' + msg; break
      case 'blue' : msg = '\u001B[34m' + msg; break
      case 'purple' : msg = '\u001B[35m' + msg; break
      default : msg = '\u001B[0m' + msg
    }
    const writer = console
    if (!a && !b && !c) {
      writer.log(common + msg)
    } else if (!b && !c) {
      writer.log(common + msg, a)
    } else if (!c) {
      writer.log(common + msg, a, b)
    } else {
      writer.log(common + msg, a, b, c)
    }
  }
}

function whatIsMyGivenContent () : any {
  return whatIam.value !== WhatIcanBe.Parent ? (props.text ? 'text' : 'no text') : innerElements.allInstanciatedElements.value
}

defineExpose<ExportedMembers>(exportedMembers)

function searchForIdealLength (originalText: string = '', targetWidth: number): string {
  let current = calculateTextWidth(originalText)

  // Now we search for the longest clipped text which fits in the target width
  if (current.width > targetWidth) {
    let maxWidth = current.width
    let maxLength = current.text.length
    let minWidth = 0
    let minLength = 0
    while (minLength < maxLength - 1) {
      totalIterationsWhenClipping++
      let averageCharWidthBetweenCurrentAndBound: number
      if (current.width > targetWidth) {
        maxLength = current.text.length
        maxWidth = current.width
        averageCharWidthBetweenCurrentAndBound = (current.width - minWidth) / (current.text.length - minLength)
      } else {
        minLength = current.text.length
        minWidth = current.width
        averageCharWidthBetweenCurrentAndBound = (maxWidth - current.width) / (maxLength - current.text.length)
      }
      // The following block estimates in `lengthToTry` the length that the clipped text should have to fulfill our target width.
      // This estimation speeds up considerably the dichotomic search by guiding it towards the optimal length. The way we use it preserves optimality.
      let estimatedLengthExcess = (current.width - targetWidth) / averageCharWidthBetweenCurrentAndBound
      if (estimatedLengthExcess > 0 && estimatedLengthExcess <= 0.5) { estimatedLengthExcess += 0.5 } // this avoids slight overflows, due to the `Math.round` just below
      estimatedLengthExcess = Math.round(estimatedLengthExcess)
      let lengthToTry = current.text.length - estimatedLengthExcess
      // If the estimation exceeds the range of possible widths, we default to a classical dichotomic search for the current iteration:
      if (lengthToTry < minLength || lengthToTry >= maxLength) {
        // ...this is why optimality is preserved.
        lengthToTry = Math.floor((minLength + maxLength) / 2)
      } else if (estimatedLengthExcess === 0) {
        break
      }
      current = calculateTextWidth(clipText(originalText, lengthToTry, numberOfEllipsesSetInProps(lengthToTry)))
    }
    numberOfClippings++
  }
  return current.text
}

function clipText (originalText: string, room: number, nEllipses : number) : string {
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
    return originalText.slice(0, midL) + '…' + originalText.slice(originalText.length - midL + r, originalText.length)
  }
  // complicated and slower
  const nBlocks = nEllipses + 1
  // First, we extract `nBlocks` blocks from the original text. The extraction algorithm guarantees that the sum of their lengths is `room`. Their lengths vary for arithmetic reasons but are as similar as possible.
  type Block = { start : number, visibleLength : number }
  const blocks : Block[] = []
  let totalToExtract = room
  let totalToSkip = originalText.length - room
  let skipNow = 0
  let s = nEllipses
  let o = 0
  for (let b = nBlocks; b > 0; b--) {
    const extractedNow = Math.round(totalToExtract / b) // do not floor here
    totalToExtract -= extractedNow
    o += skipNow
    blocks.push({ start: o, visibleLength: extractedNow })
    o += extractedNow
    skipNow = Math.floor(totalToSkip / s) // do not round here
    totalToSkip -= skipNow
    s--
  }
  // Now, we decide whether the first and/or last letter of each block should be replaced with an ellipsis. We try to create symmetry (to get something like AB…E…HI rather than AB…EF…I) because the two ends of a text are easier to catch and compare for a human eye (and we are "MiddleEllipsis" so the ellipses must look centered)
  enum Side {EndOfB, BeginningBp1}
  const sides : Side[] = [] // each element of index b indicates the side of the ellipsis between blocks b and b+1
  const symmAxis = Math.floor(nBlocks / 2)
  // first, we decide it over the first half of blocks
  let bLeft = 0
  let bRight = nBlocks - 1
  while (bLeft < symmAxis) {
    if (blocks[bLeft].visibleLength > blocks[bRight].visibleLength) { // we have a look at the opposite bloc, to try to create symmetry
      sides.push(Side.EndOfB)
      blocks[bLeft].visibleLength--
    } else {
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
    if (blocks[bRight].visibleLength > blocks[bLeft].visibleLength || blocks[bRight - 1].visibleLength <= 1) { // we have a look at the opposite bloc, trying to create symmetry with the first half while forbidding two ellipses in blocks of size 2
      sides.push(Side.BeginningBp1)
      blocks[bRight].visibleLength--
    } else {
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

function numberOfEllipsesSetInProps (textLength : number) : number {
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

const frameClassList = computed(() => 'meframe-unique000name_16218934709 ' + props.class)
</script>

<template>
  <span ref="frameSpan" :class="frameClassList">
    <!--
      The following line mounts our slot if we have one.
      To inform MiddleEllipsis components that they are children and give them the ability to send us information, we add a props. Also, we get a ref to each instanciated element.
    -->
    <component
      :is="slotElem"
      v-for="(slotElem,id) of slot"
      :key="slotElem"
      :ref="innerElements.allInstanciatedElements"
      :me-callback-to-inform-parent-about-changes="enterUpdateCycleAsAparent"
      :me-instance-id="id"
    />
  </span>
</template>

<style lang="scss">
.meframe-unique000name_16218934709 { // a fancy name is needed because we get our class list from `props.class` and we must avoid that one of those names overrides ours
  display: inline-flex;
  position: relative;
  white-space: nowrap;
  overflow: hidden;
}
</style>
