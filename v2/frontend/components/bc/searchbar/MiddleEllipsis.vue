<script setup lang="ts">
import { warn } from 'vue'

enum WhatIcanBe {
  Parent = 'parent',
  Child = 'child',
  Standalone = 'standalone',
  Error = 'oops'
}

enum WidthMode {
  NoFlexGrow,
  NoWidth,
  FixedFlexGrow,
  FixedWidth
}

type TextProperties = { text: string, width: number }

const ResizeObserverLagMargin = 1.5 // This security margin of 1.5 px is important, because the resizing observer happens to lag. If a small decrease of width making the frame as large as its content does not trigger the observer, then it will not fire anymore because the frame cannot shrink anymore.

const props = defineProps({
  text: { type: String, default: undefined },
  dontClipUnder: { type: Number, default: 0 },
  maxFlexGrow: { type: Number, default: 0 }, // use this props only if the component has no defined size (meaning that its width collapses to 0 when it contains nothing)
  middleellipsisParentGreeting: { type: Boolean, default: false } // for internal use, to inform this instance that it belongs to a parent MiddleEllipsis component
})

const slot = computed(() => { const s = useSlots(); return s.default ? s.default() : [] })

const innerElements = {
  allInstanciatedElements: ref<any[]>([]), // Instanciated elements from our slot. This array is filled by Vue in the <template>.
  // The following arrays are filled by us, each time the slot is modified:
  widthDefinedMidEllChildren: [] as any[], // instanciated elements from our slot that are MiddleEllipsis children with a defined width
  widthUndefinedMidEllChildren: [] as any[], // instanciated elements from our slot that are MiddleEllipsis children with an undefined width
  otherElements: [] as any[] // instanciated elements from our slot that are not MiddleEllipsis children
}
const frameSpan = ref<HTMLSpanElement>(null as unknown as HTMLSpanElement)

const canvasContextToCalculateTextWidths = document.createElement('canvas').getContext('2d') as CanvasRenderingContext2D
let previousTextWhoseWidthWasCalculated = ''
let previousCalculatedTextWidth = 0
let amImounted = false
let amIreadyForUpdate = false // our parent can call function getReadyForUpdate() as we can, so we use this variable to prevent multiple executions of it in a row

const whatIam = computed(() => {
  if (slot.value[0]) {
    if (props.text) {
      warn('When MiddleEllipsis is a container, it cannot receive any content in props `text`.')
      return WhatIcanBe.Error
    }
    return WhatIcanBe.Parent
  }
  if (props.text !== undefined) {
    return props.middleellipsisParentGreeting ? WhatIcanBe.Child : WhatIcanBe.Standalone
  }
  warn('MiddleEllipsis expects its props `text` to be set.')
  return WhatIcanBe.Error
})

watch(slot, () => {
  // Vue is about to fill the following array when mounting the slot in <template>.
  // We empty it, in case Vue pushes on top of the previous refs. Writing .length does not trigger Vue, which is what we want.
  innerElements.allInstanciatedElements.value.length = 0
})

watch(slot, () => { // reacts to changes of slot content (after the components of the slot are mounted by Vue in the template, thanks to `flush: 'post')
  if (whatIam.value === WhatIcanBe.Parent) {
    identifyChildren()
    updateContent()
  }
}, {
  flush: 'post' // the code above is executed after the components of the slot are mounted by Vue in the template
})

watch(() => { return props.text }, () => { // reacts to changes of text
  if (whatIam.value === WhatIcanBe.Child || whatIam.value === WhatIcanBe.Standalone) {
    updateContent()
  }
})

const resizingObserver = new ResizeObserver(() => { // will react to changes of width
  updateContent()
})

function isSizeDefined () : boolean {
  return !props.maxFlexGrow
}

onMounted(() => {
  amImounted = true
  if (whatIam.value !== WhatIcanBe.Child && whatIam.value !== WhatIcanBe.Error) {
    if (whatIam.value === WhatIcanBe.Parent) {
      identifyChildren()
    }
    resizingObserver.observe(frameSpan.value) // fires immediately...
    // ... so updateContent() is called now
  }
  // we do not do anything if we are a Child (the parent will control us)
})

onBeforeUnmount(() => {
  // Tests showed that watchers can be triggered by the unmounting cycle. We prevent any useless recalculation to improve smoothness of the UI.
  amImounted = false
  resizingObserver.disconnect()
})

function identifyChildren () {
  // the following lines refresh our information about the inner elements passed to the slot, and then we call updateContent() to manage their instances
  innerElements.widthDefinedMidEllChildren.length = 0
  innerElements.widthUndefinedMidEllChildren.length = 0
  innerElements.otherElements.length = 0
  for (const element of innerElements.allInstanciatedElements.value) {
    if (!element.updateContent || !element.isSizeDefined || !element.getReadyForUpdate) {
      // the component is not a MiddleEllipsis because it does not export those 3 functions
      innerElements.otherElements.push(element)
    } else if (element.isSizeDefined()) {
      innerElements.widthDefinedMidEllChildren.push(element)
    } else {
      innerElements.widthUndefinedMidEllChildren.push(element)
    }
  }
}

function updateContent () {
  if (whatIam.value === WhatIcanBe.Error || !amImounted || !frameSpan.value) { // in case of usage error or if we are not mounted
    return
  }
  getReadyForUpdate()
  let output = makeTextProperties(props.text)
  if (whatIam.value === WhatIcanBe.Child || whatIam.value === WhatIcanBe.Standalone) { // if we are meant to shorten and display a text
    if (output.text.length >= props.dontClipUnder) {
      output = searchForIdealLength(output.text, getFrameWidth() - ResizeObserverLagMargin)
    }
  } else if (whatIam.value === WhatIcanBe.Parent) { // if we are meant to manage children
    // we ask all children to get ready for an update (this includes finding an initial (and temporary) width for those having an undefined width)
    for (const child of innerElements.widthDefinedMidEllChildren) { child.getReadyForUpdate() }
    for (const child of innerElements.widthUndefinedMidEllChildren) { child.getReadyForUpdate() }
    // first we allow children with an undefined width to update their content
    for (const child of innerElements.widthUndefinedMidEllChildren) { child.updateContent() }
    // now that the undefined ones got a width, we allow the others to use the remaining room (this makes sense when their width is loosely defined with flex-grow)
    for (const child of innerElements.widthDefinedMidEllChildren) { child.updateContent() }
  }
  settleAfterUpdate(output)
}

function searchForIdealLength (originalText : string, targetWidth : number) : TextProperties {
  let current = makeTextProperties(originalText)

  // Now we search for the longest clipped text which fits in the target width
  if (current.width > targetWidth) {
    let maxWidth = current.width
    let maxLength = current.text.length
    let minWidth = 0
    let minLength = 0
    while (minLength < maxLength - 1) {
      let averageCharWidthBetweenCurrentAndBound : number

      if (current.width > targetWidth) {
        maxLength = current.text.length
        maxWidth = current.width
        averageCharWidthBetweenCurrentAndBound = (current.width - minWidth) / (current.text.length - minLength)
      } else {
        minLength = current.text.length
        minWidth = current.width
        averageCharWidthBetweenCurrentAndBound = (maxWidth - current.width) / (maxLength - current.text.length)
      }

      // this estimation speeds up considerably the dichotomic search by guiding it progressively towards the optimal length
      let estimatedLengthExcess = (current.width - targetWidth) / averageCharWidthBetweenCurrentAndBound
      if (estimatedLengthExcess > 0 && estimatedLengthExcess <= 0.5) { estimatedLengthExcess += 0.5 } // this avoids slight overflows, due to the Math.round just below
      estimatedLengthExcess = Math.round(estimatedLengthExcess)
      let lengthToTry = current.text.length - estimatedLengthExcess

      if (lengthToTry < minLength || lengthToTry >= maxLength) {
        // if the estimation exceeds the range of possible widths, we default to a classical search by dichotomy *for this iteration*
        lengthToTry = Math.floor((minLength + maxLength) / 2)
      } else if (estimatedLengthExcess === 0) {
        break
      }
      current = makeTextProperties(shortenText(originalText, lengthToTry))
    }
  }
  return current
}

function shortenText (originalText : string, maxLength : number) : string {
  if (originalText.length <= maxLength) {
    return originalText
  }
  if (maxLength <= 0) {
    return ''
  }
  if (maxLength === 1) {
    return '…'
  }

  const midL = Math.floor(maxLength / 2)
  const roomForEllipsis = 1 - (maxLength % 2)

  return originalText.slice(0, midL - roomForEllipsis) + '…' + originalText.slice(originalText.length - midL, originalText.length)
}

function prepareTextWidthCalculations () {
  previousTextWhoseWidthWasCalculated = ''
  previousCalculatedTextWidth = 0
  if (frameSpan.value) {
    canvasContextToCalculateTextWidths.font = getComputedStyle(frameSpan.value).font
  }
}

function calculateTextWidth (text : string) : number {
  if (text === previousTextWhoseWidthWasCalculated) {
    // This is intended to speed up the beggining of searchForIdealLength(). At the beginning of the function, the width of the original
    // text is requested but the width has already been calculated just before, when updateContent() called makeTextProperties()
    return previousCalculatedTextWidth
  }
  previousTextWhoseWidthWasCalculated = text
  previousCalculatedTextWidth = canvasContextToCalculateTextWidths.measureText(text).width
  return previousCalculatedTextWidth
}

// Caution: prepareTextWidthCalculations() must be called at some point before
function makeTextProperties (text : string | undefined) : TextProperties {
  if (!text) {
    return { text: '', width: 0 }
  }
  return { text, width: calculateTextWidth(text) }
}

function setFrameText (text : string) {
  if (frameSpan.value) {
    frameSpan.value.textContent = text
  }
}

function getFrameWidth () : number {
  if (!frameSpan.value) {
    return 0
  }
  return frameSpan.value.clientWidth
}

function setFrameWidth (mode : WidthMode, x : number = 0) {
  if (!frameSpan.value) {
    return
  }
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
      frameSpan.value.style.setProperty('width', String(x) + 'px')
    }
  }
}

function getReadyForUpdate () {
  if (amIreadyForUpdate) {
    // setFrameWidth() and prepareTextWidthCalculations() are slow functions (each of them triggers a reflow) so we do nothing if getReadyForUpdate() has already been executed (amIreadyForUpdate is true)
    return
  }
  amIreadyForUpdate = true
  if (!isSizeDefined()) {
    setFrameWidth(WidthMode.FixedFlexGrow, props.maxFlexGrow) // This is why a MiddleEllipsis cannot have an undefined width unless it is inside a Parent. This line would trigger the resize observer so create an infinite loop of updates.
  }
  if (whatIam.value !== WhatIcanBe.Parent) {
    setFrameText('')
    prepareTextWidthCalculations()
  }
}

function settleAfterUpdate (finalText : TextProperties) {
  if (whatIam.value !== WhatIcanBe.Parent) {
    setFrameText(finalText.text)
  }
  if (!isSizeDefined()) {
    setFrameWidth(WidthMode.NoFlexGrow) // This is why a MiddleEllipsis cannot have an undefined width unless it is inside a Parent. This line would trigger the resize observer so create an infinite loop of updates.
  }
  amIreadyForUpdate = false
}

defineExpose({ // for the parent to control this component as a child
  updateContent,
  isSizeDefined,
  getReadyForUpdate
})
</script>

<template>
  <span ref="frameSpan" class="frame">
    <!--
      The following mounts our slot if we have one.
      To inform MiddleEllipsis components that they are children, we add a props. Also, we get a ref to each instanciated element.
    -->
    <component
      :is="slotElem"
      v-for="slotElem of slot"
      :key="slotElem"
      :ref="innerElements.allInstanciatedElements"
      :middleellipsis-parent-greeting="true"
    />
  </span>
</template>

<style lang="scss" scoped>
.frame {
  display: inline-flex;
  position: relative;
  white-space: nowrap;
  overflow: clip;
}
</style>
