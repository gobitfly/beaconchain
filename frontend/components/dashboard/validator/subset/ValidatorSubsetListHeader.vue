<script lang="ts" setup>
import {
  faDesktop,
  faPowerOff,
  faMoneyBill
} from '@fortawesome/pro-solid-svg-icons'
import {
  faClock,
  type IconDefinition
} from '@fortawesome/pro-regular-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import type { ValidatorSubsetCategory } from '~/types/validator'
import type { VDBSummaryValidator } from '~/types/api/validator_dashboard'
import type { SlotVizCategories } from '~/types/dashboard/slotViz'

interface Props {
  category: ValidatorSubsetCategory,
  validators: VDBSummaryValidator[],
}
const props = defineProps<Props>()

const { t: $t } = useI18n()

const icon = computed(() => {
  let icon: IconDefinition | undefined
  let className = ''
  let slotVizCategory: SlotVizCategories | undefined
  switch (props.category) {
    case 'deposited':
      icon = faMoneyBill
      break
    case 'offline':
      className = 'negative'
      icon = faPowerOff
      break
    case 'online':
      className = 'positive'
      icon = faPowerOff
      break
    case 'pending':
      icon = faClock
      break
    case 'sync_current':
      className = 'positive'
      slotVizCategory = 'sync'
      break
    case 'sync_upcoming':
      className = 'positive'
      slotVizCategory = 'sync'
      break
    case 'sync_past':
      className = 'text-disabled'
      slotVizCategory = 'sync'
      break
    case 'has_slashed':
      className = 'positive'
      slotVizCategory = 'slashing'
      break
    case 'got_slashed':
      className = 'negative'
      slotVizCategory = 'slashing'
      break
    case 'proposal_proposed':
      className = 'positive'
      slotVizCategory = 'proposal'
      break
    case 'proposal_missed':
      className = 'negative'
      slotVizCategory = 'proposal'
      break
    default:
      icon = faDesktop
      break
  }

  return { icon, className, slotVizCategory }
})

</script>

<template>
  <div class="subset--list-header">
    <FontAwesomeIcon v-if="icon.icon" :icon="icon.icon" :class="icon.className" />
    <SlotVizIcon v-else-if="icon.slotVizCategory" :icon="icon.slotVizCategory" :class="icon.className" />
    <span>{{ $t(`dashboard.validator.subset_dialog.category.${category}`) }}</span>
    <span> ({{ validators.length }})</span>
  </div>
</template>

<style lang="scss" scoped>
.subset--list-header {
  display: flex;
  align-items: center;
  gap: var(--padding);
}
</style>
