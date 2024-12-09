<script setup lang="ts" generic="T">
import { faCopy } from '@fortawesome/pro-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

const props = defineProps<{
  infoCopy?: string,
  item?: T,
  items?: T[],
  open?: boolean,
}>()
const isOpen = ref(props.open ?? false)
const {
  copy,
  isSupported,
} = useClipboard()
const idList = useId()

const textToCopy = ref<string>('')
const detailElement = ref<HTMLElement>()
onMounted(() => {
  if (!detailElement.value) return
  const ulElement = detailElement.value.querySelector(`#${idList}`)
  const liElements = ulElement?.querySelectorAll('li')
  if (liElements) {
    const liTexts = [ ...liElements ].map(liElement => liElement.textContent?.trim())
    textToCopy.value = liTexts.join(', ')
  }
})
const toast = useBcToast()
const { t: $t } = useTranslation()
const copyText = async () => {
  await copy(textToCopy.value).then(() => {
    toast.showInfo({
      detail: $t('clipboard.copied_to_clipboard'),
      summary: props.infoCopy ?? '',
    })
  })
}
</script>

<template>
  <details
    ref="detailElement"
    class="bc-accordion"
    :open
  >
    <summary
      @click="isOpen = !isOpen"
    >
      <span class="bc-accordion__heading">
        <IconChevron
          :direction="isOpen ? 'bottom' : 'right'"
        />
        <slot name="headingIcon" />
        <slot name="heading" />
      </span>
    </summary>
    <BcCard class="bc-accordion__content">
      <ul
        v-if="items?.length"
        :id="idList"
        class="bc-accordion-list"
      >
        <li
          v-for="(element, index) in items"
          :key="`${index}-${element}`"
          class="bc-accordion-list__element"
        >
          <slot
            name="item"
            :item="element"
          />
        </li>
      </ul>
      <slot
        v-if="item"
        name="item"
        :item
      />
      <template #floating-action-button>
        <BcButtonIcon
          v-if="isSupported"
          screenreader-text="Copy list to clipboard"
          class="bc-accordion__button"
          @click="copyText"
        >
          <FontAwesomeIcon
            :icon="faCopy"
            class="bc-accordion__button-icon"
          />
        </BcButtonIcon>
      </template>
    </BcCard>
  </details>
</template>

<style scoped lang="scss">
@use '~/assets/css/breakpoints' as *;

.bc-accordion {
  position: relative;
  summary {
    list-style: none;
    cursor: pointer;
  }
  summary::-webkit-details-marker{
    display: none;
  }
}
.bc-accordion__heading {
  display: inline-flex;
  align-items: center;
  gap: 0.625rem;
}
.bc-accordion__content {
  margin-top: 0.625rem;
  min-height: 2.563rem;
  max-height: 9.375rem;
  overflow: auto;
  width: 100%;
  @media (min-width: $breakpoint-md) {
    width: 41.75rem;
  }

}
.bc-accordion-list {
  list-style: none;
  padding-inline-start: 0;
  display: inline;
}
.bc-accordion-list__element {
  display: inline;
}
.bc-accordion-list__element:not(:last-child)::after {
content: ', ';
}
.bc-accordion__button {
  --border-width: .0625rem;
  background-color: var(--input-background);
  border-radius: var(--corner-radius, 4px);
  padding: calc(0.3125rem - var(--border-width));
  border: var(--border-width) solid var(--input-border-color);
  color: inherit;
  height: 1.875rem;
  width: 1.875rem;
  position: absolute;
  bottom: 6px;
  right: 11px;
  cursor: pointer
}
.bc-accordion__button-icon {
  width: .9375rem;
  font-size: .9375rem;
  line-height: 100%;
}
</style>
