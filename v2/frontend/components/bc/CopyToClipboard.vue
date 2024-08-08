<script setup lang="ts">
import { faCopy } from '@fortawesome/pro-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { warn } from 'vue'
import BcTooltip from './BcTooltip.vue'

interface Props {
  value?: string
}
const props = defineProps<Props>()

const { t: $t } = useTranslation()
const {
  bounce,
  instant,
  value: tooltip,
} = useDebounceValue<string>($t('clipboard.copy'), 2000)

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
      bounce($t('clipboard.copy'))
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
    <FontAwesomeIcon
      :icon="faCopy"
      class="pointer"
      @click.stop.prevent="copyToClipboard"
    />
  </BcTooltip>
</template>

<style>
.tooltip {
  min-width: max-content;
}
</style>
