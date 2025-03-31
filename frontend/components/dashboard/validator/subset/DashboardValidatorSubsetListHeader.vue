<script lang="ts" setup>
import type { ValidatorSubsetCategory } from '~/types/validator'
import type { VDBSummaryValidator } from '~/types/api/validator_dashboard'
import type { SlotVizCategories } from '~/types/dashboard/slotViz'
import { countSummaryValidatorDuties } from '~/utils/dashboard/validator'
import type { Icon } from '~/components/bc/icon/BcIcon.vue'

interface Props {
  category: ValidatorSubsetCategory,
  validators: VDBSummaryValidator[],
}
const props = defineProps<Props>()

const { t: $t } = useTranslation()

const icon = computed(() => {
  let icon: Icon | undefined
  let className = ''
  let slotVizCategory: SlotVizCategories | undefined
  switch (props.category) {
    case 'deposited':
    case 'exited':
    case 'exited_withdrawing':
    case 'exited_withdrawn':
      icon = 'money-bill'
      break
    case 'got_slashed':
      className = 'negative'
      slotVizCategory = 'slashing'
      break
    case 'has_slashed':
      className = 'positive'
      slotVizCategory = 'slashing'
      break
    case 'offline':
      className = 'negative'
      icon = 'power-off'
      break
    case 'online':
      className = 'positive'
      icon = 'power-off'
      break
    case 'pending':
      icon = 'clock'
      break
    case 'proposal_missed':
      className = 'negative'
      slotVizCategory = 'proposal'
      break
    case 'proposal_proposed':
      className = 'positive'
      slotVizCategory = 'proposal'
      break
    case 'slashed':
    case 'slashed_withdrawing':
      slotVizCategory = 'slashing'
      break
    case 'sync_current':
      className = 'positive'
      slotVizCategory = 'sync'
      break
    case 'sync_past':
      className = 'text-disabled'
      slotVizCategory = 'sync'
      break
    case 'sync_upcoming':
      className = 'positive'
      slotVizCategory = 'sync'
      break
    default:
      icon = 'desktop'
      break
  }

  return {
    className,
    icon,
    slotVizCategory,
  }
})

const count = computed(() =>
  countSummaryValidatorDuties(props.validators, props.category),
)
</script>

<template>
  <div class="subset--list-header">
    <BcIcon
      v-if="icon.icon"
      :name="icon.icon"
      :class="icon.className"
    />
    <DashboardSlotVizDutyIcon
      v-else-if="icon.slotVizCategory"
      :icon="icon.slotVizCategory"
      :class="icon.className"
    />
    <span>{{
      $t(`dashboard.validator.subset_dialog.category.${category}`)
    }}</span>
    <span> (<BcFormatNumber :value="count" />)</span>
  </div>
</template>

<style lang="scss" scoped>
.subset--list-header {
  display: flex;
  align-items: center;
  gap: var(--padding);
}
</style>
