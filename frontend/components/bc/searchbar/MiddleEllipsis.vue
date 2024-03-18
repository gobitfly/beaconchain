<script setup lang="ts">
/*
  Usage:
*/

/*
  Because the text is not shortened by the CSS engine, we must search for the text length that fits the best the container without overflowing.
  This search involves trials and errors: several iterations with different text lengths are done for every instance of this component on the page.
  Therefore we must absolutely find the right length as quickly as possible and try the different lengths without causing flikering nor blurry effects.
  Here is the strategy that I suggest to fulfill those constraints:
  1a. If the parent has not fixed a width, let the browser write in a span the full text passed in the slot and, if it overflows, let it clip it and set the component width
      with the official rules of HTML&CSS.
  1b. Or, if the parent signals that it has fixed a width, empty the span. We empty the span because "fixing a width" can be done loosely:
      When the content overflows and the width has been fixed loosely by the parent (for example with `flex-grow`), the component might still grow larger than its container.
  2.  Measure the size of the component.
  3.  Make the content invisible to avoid flickering and blurry effects during the search.
  4.  Run a dichotomic search (in O(log n)) in the span, to find the largest text that we can fit within the ideal size that we measured.
      Guide the search by influencing it with the approximate size of the text that might fit, calculated by combining the width of the text, the text length and the component width.
      This guidance speeds up significantly the search: my tests (hashes in the search bar) show that we iterate 3 times on average versus 7 times with a pure dichotomy.
  5.  Once the optimal text size in found, remove transparency.
*/

enum FrameWidthMode {
  AdaptiveInParent = 'frame-adaptive-in-parent',
  FixedInParent = 'frame-fixed-in-parent'
}

const props = defineProps({ widthIsFixed: { type: Boolean, default: false } })
const frameSpan = ref<HTMLSpanElement>()
const contentSpan = ref<HTMLSpanElement>()
const contentVisibility = ref<string>('hidden')
let originalText = ''
const frameWidthMode = props.widthIsFixed ? FrameWidthMode.FixedInParent : FrameWidthMode.AdaptiveInParent

onMounted(() => {
  originalText = getSpanText()
  searchForIdealLength()
})

function searchForIdealLength () {
  setSpanVisibility(false)

  // The following lines measure the maximum width that the component is authorized to take and store this information in `targetWidth`
  if (frameWidthMode === FrameWidthMode.FixedInParent) {
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

.frame-adaptive-in-parent {
  max-width: 100%;
}

.frame-fixed-in-parent {
  width: 100%;
}

.content {
  display: inline-block;
  position: relative;
  white-space: nowrap;
  overflow: clip;
  visibility: v-bind(contentVisibility);
}
</style>
