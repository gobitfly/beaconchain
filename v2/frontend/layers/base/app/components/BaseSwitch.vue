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
const id = useId()
const name = useId()
const modelValue = defineModel<typeof values[number]['key']>({ required: true })
const defaultValue = modelValue.value
const { t: $t } = useTranslation()
const thumbs = useTemplateRef('thumb')
const track = useTemplateRef('track')

const moveThumb = () => {
  const activeTrackItem = track.value?.querySelector(':has(input[type="radio"]:checked)')
  const thumb = thumbs.value?.[0]
  if (!activeTrackItem) return
  if (!thumb) return

  const { x: initialX } = thumb.getBoundingClientRect()
  const {
    width,
    x,
  } = activeTrackItem.getBoundingClientRect()

  // this should rather have been done via view transition api
  // but it currently lacks `firefox support`
  // and also there was a flickering issue
  const animation = thumb.animate([ {
    transform: `translateX(${x - initialX}px)`,
    width: `${width}px`,
  } ],
  { duration: 180 },
  )
  return animation.finished.then(() => {
    if (!thumb) return
    activeTrackItem.appendChild(thumb)
  })
}

watch(modelValue, () => {
  moveThumb()
})
</script>

<template>
  <fieldset
    class="isolate"
  >
    <legend class="sr-only">
      {{ $t(screenreaderTitle as string) }}
    </legend>
    <div
      ref="track"
      :class="classList?.track"
      v-bind="$attrs"
    >
      <label
        v-for="(value, index) in values"
        :key="value.key"
        :for="`${id}-${index}`"
        :class="classList?.trackItem"
        class="relative"
      >
        <span class="relative z-10">
          <slot
            :name="value.key"
            :label="values[index]?.label"
          >
            {{ values[index]?.label }}
          </slot>
        </span>
        <input
          :id="`${id}-${index}`"
          v-model="modelValue"
          :value="values[index]?.key"
          type="radio"
          class="sr-only"
          :name
        >
        <span
          v-if="value.key === defaultValue"
          ref="thumb"
          class="absolute inset-[0] z-0"
          aria-hidden="true"
          :class="classList?.thumb"
        />
      </label>
    </div>
  </fieldset>
</template>

<style scoped></style>
