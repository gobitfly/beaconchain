<script setup lang="ts">
/*
  Usage:
*/

/*
  Because the text is not shortened by the CSS engine, we must search for the text length that fits the best the container without overflowing.
  This search involves trials and errors: different text lengths are tried for every instance of this component on the page.
  Therefore we must absolutely find the right length as quickly as possible and try the different lengths without causing flikering nor blurry effects with the component and its neighbors.
  Here is the strategy that I suggest to fulfill those constraints:
  1a. If the parent has not fixed a width, let the browser write in a span the full text passed in the slot and, if it overflows, let it clip it after it set the component width
      with the official rules of HTML&CSS.
  1b. Or, if the parent signals that it has fixed the width, empty the span. We empty the span because "fixing a width" can be done loosely:
      when the content overflows and the width has been fixed loosely by the parent (for example with `flex-grow`), the component might still grow larger than its container.
  2.  Measure the width of the component: this is our target width. We want to find the longest text possible within the target.
  3a. Make the content invisible to avoid flickering and blurry effects during the search.
  3b. Force the component width to the target width. This makes sure that the neighbor components will not be pulled and pushed repetitively while we try different text lengths (that also speeds us up).
  4.  Run a dichotomic search (in O(log n)) in the span, to find the largest text that we can fit within the target.
      Guide the search by influencing it with the approximate length of the text that might fit, calculated by combining the width of the text, the text length and the component width.
      This guidance speeds up significantly the search: my tests (hashes in the search bar) show that we iterate 3 times on average versus 7 times with a pure dichotomy.
      Of course, if the original text is smaller than the target, 0 iteration happens.
  5.  Unfix the component width to recover its original dynamic settings and make the component visible.
*/

enum FrameWidthMode {
  AdaptiveInParent = 'frame-adaptivewidth-in-parent',
  FixedInParent = 'frame-fixedwidth-in-parent',
  FixedHere = 'frame-forcedwidth-here'
}

const props = defineProps({ widthIsFixed: { type: Boolean, default: false } })
const frameSpan = ref<HTMLSpanElement>()
const contentSpan = ref<HTMLSpanElement>()

const contentVisibility = ref<string>('hidden')
const frameWidthMode = ref<FrameWidthMode>(getOriginalFrameWidthMode())
let frameWidthIfForced = ''

let originalText = ''

onMounted(() => {
  originalText = getSpanText()
  searchForIdealLength()
})

function searchForIdealLength () {
  setSpanVisibility(false)

  // The following lines measure the maximum width that the component is authorized to take and store this information in `targetWidth`
  if (frameWidthMode.value === FrameWidthMode.FixedInParent) {
    setSpanText('') // we do this to make sure that we will measure the width desired by the parent, the component does not grow larger than that.
  } else {
    setSpanText(originalText) // if the parent signals that it did not fix a width, we leave the text in the span to let the browser find a width following HTML and CSS rules (the parent must have set a max-width)
  }
  const targetWidth = getSpanWidth(frameSpan)

  // now we measure the width of the span when it contains the complete text
  setSpanText(originalText)
  let maxWidth = getSpanWidth(contentSpan)
  let maxLength = originalText.length

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
      setSpanText(shortenText(lengthToTry))
    }

    setFrameWidth(undefined)
  }

  setSpanVisibility(true)
}

function shortenText (maxLength : number) : string {
  if (originalText.length <= maxLength) {
    return originalText
  }

  const midL = Math.floor(maxLength / 2)
  const roomForEllipsis = 1 - (maxLength % 2)

  return originalText.slice(0, midL - roomForEllipsis) + '…' + originalText.slice(originalText.length - midL, originalText.length)
}

function getSpanText () : string {
  if (contentSpan.value === undefined || contentSpan.value.textContent === null) {
    return ''
  }
  return contentSpan.value.textContent
}

function setSpanText (text : string) {
  if (contentSpan.value !== undefined) {
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

.frame-adaptivewidth-in-parent {
  max-width: 100%;
}

.frame-fixedwidth-in-parent {
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
