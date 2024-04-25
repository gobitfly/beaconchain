<script lang="ts" setup>
import {
  faGem
} from '@fortawesome/pro-regular-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { BcPremiumModal } from '#components'

interface Props {
  description?: string,
  dismissLabel?: string
}
const props = defineProps<Props>()

const dialog = useDialog()
const { t: $t } = useI18n()

const openPremiumModal = () => {
  dialog.open(BcPremiumModal, {
    data: {
      description: props.description, dismissLabel: props.dismissLabel
    }
  })
}
</script>

<template>
  <BcTooltip :text="$t('premium.subscribe')">
    <div @click.stop.prevent="openPremiumModal()">
      <FontAwesomeIcon :icon="faGem" class="gem" />
    </div>
  </BcTooltip>
</template>

<style lang="scss" scoped>
.gem {
  color: var(--primary-color);
  cursor: pointer;
}
</style>
