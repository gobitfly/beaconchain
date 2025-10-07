<script setup lang="ts" generic="T extends SwitchValue[]">
export type SwitchValue = {
  key: string,
  label: string,
}
const { values } = defineProps<{
  classList?: {
    thumb?: string,
    track?: string,
    trackItem?: string,
  },
  screenreaderTitle: TranslationInput,
  values: T,
}>()
const idFirstValue = useId()
const idSecondValue = useId()
const name = useId()
const modelValue = defineModel<typeof values[number]['key']>()
const { t: $t } = useTranslation()
const thumb = useTemplateRef('thumb')
const track = useTemplateRef('track')

const moveThumb = () => {
  const activeTrackItem = track.value?.querySelector(':has(input[type="radio"]:checked)')
  if (!activeTrackItem) return
  if (!thumb.value) return

  const { x: initialX } = thumb.value.getBoundingClientRect()
  const { x } = activeTrackItem.getBoundingClientRect()

  // this should rather have been done via view transition api
  // but it currently lacks `firefox support`
  // and also there was a flickering issue
  const animation = thumb.value.animate([ {
    transform: `translateX(${x - initialX}px)`,
  } ],
  { duration: 180 },
  )
  return animation.finished.then(() => {
    activeTrackItem.appendChild(thumb.value!)
  })
}
watch(modelValue, () => {
  moveThumb()
})
</script>

<template>
  <fieldset class="isolate">
    <legend class="sr-only">
      {{ $t(screenreaderTitle as string) }}
    </legend>
    <div
      ref="track"
      :class="classList?.track"
      v-bind="$attrs"
    >
      <label
        :for="idFirstValue"
        :class="classList?.trackItem"
        class="relative"
      >
        <span
          class="relative z-10"
        >
          {{ values[0]?.label }}
        </span>
        <input
          :id="idFirstValue"
          v-model="modelValue"
          :value="values[0]?.key"
          type="radio"
          class="appearance-none"
          :name
        >
        <span
          ref="thumb"
          class="absolute inset-[0] z-0"
          aria-hidden="true"
          :class="classList?.thumb"
        />
      </label>
      <label
        class="relative"
        :class="classList?.trackItem"
        :for="idSecondValue"
      >
        <span
          class="relative z-10"
        >
          {{ values[1]?.label }}
        </span>
        <input
          :id="idSecondValue"
          v-model="modelValue"
          :value="values[1]?.key"
          class="appearance-none"
          type="radio"
          :name
        >
      </label>
    </div>
  </fieldset>
</template>

<style scoped></style>
