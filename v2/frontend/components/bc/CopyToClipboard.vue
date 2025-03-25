<script setup lang="ts">
import { warn } from 'vue'
import BcTooltip from './BcTooltip.vue'

interface Props {
  value?: string,
}
const props = defineProps<Props>()

const { t: $t } = useTranslation()
const {
  bounce,
  instant,
  value: tooltip,
} = useDebounceValue<string>($t('clipboard.action.copy_to_clipboard'), 2000)

function copyToClipboard(): void {
  if (!props.value) {
    return
  }

  navigator.clipboard
    .writeText(props.value)
    .catch((error) => {
      warn('Error copying text to clipboard:', error)
    })
    .then(() => {
      instant($t('clipboard.copied'))
      bounce($t('clipboard.action.copy_to_clipboard'))
    })
}
</script>

<template>
  <BcTooltip
    v-if="props.value"
    :text="tooltip"
    position="top"
    tooltip-class="tooltip"
  >
    <BcButtonIcon
      name="copy"
      class="pointer"
      screenreader-text="clipboard.action.copy_to_clipboard"
      @click.stop.prevent="copyToClipboard"
    />
  </BcTooltip>
</template>

<style>
.tooltip {
  min-width: max-content;
}
</style>
