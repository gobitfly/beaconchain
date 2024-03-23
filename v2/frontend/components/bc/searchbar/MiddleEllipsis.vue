<script setup lang="ts">
/*
  This component clips the text that you give in its slot. The text is clipped in the middle so the beginning and the end of the text
  remain visible.

  Use it with the following syntax:
  <BcSearchbarMiddleEllipsis>your long text</BcSearchbarMiddleEllipsis>

  The slot cannot contain HTML and components. The slot must contain text only. The text can be generated at run time with {{ }}
  Therefore, to style the text, assign a class to the component or to its parent container.
*/

/*
  Internal functionning:
  CSS allows to clip text only at its end, not in the middle. So this component searches for the text length that fits the best the container
  without overflowing.
  This search involves trials and errors: different text lengths are tried for every instance of this component on the page.
  Therefore we must absolutely do it as quickly as possible and the different attempts must not cause flikering nor blurry effects with the
  component as well as its neighbors.
  Here is the strategy that I suggest to fulfill those constraints:
  1.  Make the content invisible to avoid flickering and blurry effects during the procedure.
  2a. If the parent has not fixed a width, let the browser write in a span the full text passed in the slot and, if it overflows, let it clip it
      after it set the component width with the official rules of HTML&CSS. This is not the clipping style that we want but it tells us which width
      the component must have.
  2b. Or, if the parent signals that it has fixed the width, empty the span. We empty the span because "fixing a width" might have been done loosely:
      if the content overflows and the width has been fixed loosely by the parent (for example with `flex-grow`), the component might still grow
      larger than its container.
  3.  Now that either 2a or 2b is done, measure the width of the component: this is our target width. We will find the longest text possible
      within that target.
  4.  Force the component width to the target width. This makes sure that the neighbor components will not be pulled and pushed repetitively while
      we try different text lengths (that also speeds us up).
  5.  Run a dichotomic search (in O(log n)) in the span, to find the largest text that we can fit within the target.
      Guide the search by influencing it with the approximate length of the text that might fit, calculated by combining the width of the text, the
      text length and the component width. This guidance speeds up significantly the search: my tests (hashes in the search bar) show that we
      iterate 3 times on average versus 7 times with a pure dichotomy.
      Of course, if the original text is smaller than the target, 0 iteration happens.
  6.  Unfix the component width to recover its original setting and make the content visible.
*/

enum FrameWidthMode {
  AdaptiveInParent = 'frame-adaptivewidth-inparent',
  FixedInParent = 'frame-fixedwidth-inparent',
  FixedHere = 'frame-forcedwidth-here'
}

const props = defineProps({ widthIsFixed: { type: Boolean, default: false } })
const frameSpan = ref<HTMLSpanElement>()
const contentSpan = ref<HTMLSpanElement>()
const defaultSlot = useSlots().default

const contentVisibility = ref<string>('hidden')
const frameWidthMode = ref<FrameWidthMode>(getOriginalFrameWidthMode())
let frameWidthIfForced = ''

onMounted(() => { // reacts when the component is displayed on the client
  updateShortenedText()
})
watch(() => { return defaultSlot }, () => { // reacts to changes of slot content
  updateShortenedText()
})
const resizingObserver = new ResizeObserver(updateShortenedText) // will react to changes of component width

function updateShortenedText () {
  if (frameSpan.value !== undefined && contentSpan.value !== undefined && contentSpan.value !== null) {
    const originalText = (defaultSlot === undefined) ? '' : String(defaultSlot()[0].children)
    resizingObserver.unobserve(frameSpan.value) // makes sure that calls will not be triggerred by the different text lengths that searchForIdealLength() will try
    searchForIdealLength(originalText)
    resizingObserver.observe(frameSpan.value) // makes sure that the text remains ideally shortened when the component gets more or less room
  }
}

function searchForIdealLength (originalText : string) {
  if (originalText === '') {
    setSpanText('')
    return
  }

  setSpanVisibility(false)

  // The following paragraph measures the maximum width that the component is authorized to take and stores this information in `targetWidth`
  if (frameWidthMode.value === FrameWidthMode.FixedInParent) {
    setSpanText('') // we do this to make sure that we will measure the width desired by the parent, the component does not grow larger than that.
  } else {
    setSpanText(originalText) // if the parent signals that it did not fix a width, we leave the text in the span to let the browser find a width following HTML and CSS rules (the parent must have set a max-width)
  }
  const targetWidth = getSpanWidth(frameSpan)

  // This paragraph measures the width of the span when it contains the complete text
  setSpanText(originalText)
  let maxWidth = getSpanWidth(contentSpan)
  let maxLength = originalText.length

  // Now we search for the longest clipped text which fits in the target width
  if (maxWidth > targetWidth) {
    setFrameWidth(targetWidth) // to avoid pulling and pushing repetitively neighbor components while we try different text lengths
    let minWidth = 0
    let minLength = 0

    while (minLength < maxLength - 1) {
      let averageCharWidthBetweenCurrentAndBound : number

      if (getSpanWidth(contentSpan) > targetWidth) {
        maxLength = getSpanText().length
        maxWidth = getSpanWidth(contentSpan)
        averageCharWidthBetweenCurrentAndBound = (getSpanWidth(contentSpan) - minWidth) / (getSpanText().length - minLength)
      } else {
        minLength = getSpanText().length
        minWidth = getSpanWidth(contentSpan)
        averageCharWidthBetweenCurrentAndBound = (maxWidth - getSpanWidth(contentSpan)) / (maxLength - getSpanText().length)
      }

      const estimatedLengthGap = Math.round((getSpanWidth(contentSpan) - targetWidth) / averageCharWidthBetweenCurrentAndBound)
      let lengthToTry = getSpanText().length - estimatedLengthGap

      if (lengthToTry < minLength || lengthToTry >= maxLength) {
        // if the estimation exceeds the range of possible widths, we default to a classical search by dichotomy *for this iteration*
        lengthToTry = Math.floor((minLength + maxLength) / 2)
      } else if (estimatedLengthGap === 0) {
        break
      }
      setSpanText(shortenText(originalText, lengthToTry))
    }

    setFrameWidth(undefined)
  }

  setSpanVisibility(true)
}

function shortenText (originalText : string, maxLength : number) : string {
  if (originalText.length <= maxLength) {
    return originalText
  }
  if (maxLength <= 0) {
    return ''
  }

  const midL = Math.floor(maxLength / 2)
  const roomForEllipsis = 1 - (maxLength % 2)

  return originalText.slice(0, midL - roomForEllipsis) + '…' + originalText.slice(originalText.length - midL, originalText.length)
}

function getSpanText () : string {
  if (contentSpan.value === undefined || contentSpan.value === null || contentSpan.value.textContent === null) {
    return ''
  }
  return contentSpan.value.textContent
}

function setSpanText (text : string) {
  if (contentSpan.value !== undefined && contentSpan.value !== null) {
    contentSpan.value.textContent = text
  }
}

function setSpanVisibility (visible : boolean) {
  contentVisibility.value = visible ? 'visible' : 'hidden'
}

function getSpanWidth (whichOne : Ref<HTMLSpanElement | undefined>) : number {
  if (whichOne.value === undefined) {
    return 0
  }
  return whichOne.value.clientWidth
}

function setFrameWidth (size : number | undefined) {
  if (size === undefined) {
    frameWidthMode.value = getOriginalFrameWidthMode()
  } else {
    frameWidthIfForced = String(size) + 'px'
    frameWidthMode.value = FrameWidthMode.FixedHere
  }
}

function getOriginalFrameWidthMode () : FrameWidthMode {
  return props.widthIsFixed ? FrameWidthMode.FixedInParent : FrameWidthMode.AdaptiveInParent
}
</script>

<template>
  <span ref="frameSpan" class="frame" :class="frameWidthMode">
    <span ref="contentSpan" class="content"><slot /></span>
  </span>
</template>

<style lang="scss" scoped>
.frame {
  display: inline-block;
  position: relative;
  overflow: clip;
}

.frame-adaptivewidth-inparent {
  max-width: 100%;
}

.frame-fixedwidth-inparent {
  width: 100%;
}

.frame-forcedwidth-here {
  width: v-bind(frameWidthIfForced)
}

.content {
  display: inline-block;
  position: relative;
  white-space: nowrap;
  overflow: clip;
  visibility: v-bind(contentVisibility);
}
</style>
