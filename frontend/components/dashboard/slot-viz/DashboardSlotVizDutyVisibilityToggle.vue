<script setup lang="ts">
import { faEye } from '@fortawesome/pro-solid-svg-icons'
import type { MultiBarItem } from '~/types/multiBar'
import {
  IconSlotAttestation,
  IconSlotBlockProposal,
  IconSlotSlashing,
  IconSlotSync,
} from '#components'
import type { SlotVizCategories } from '~/types/dashboard/slotViz'
import { COOKIE_KEY } from '~/types/cookie'

interface Props {
  initiallyHideVisible?: boolean,
}
const props = defineProps<Props>()

type SlotVizCategoriesStorage = {
  [dashboardId: string]: SlotVizCategories[],
}

const { t: $t } = useTranslation()
const { dashboardKey } = useDashboardKey()

const persistedSlotVizCategories = useCookie<SlotVizCategoriesStorage>('selected-slotviz-categories', { default: () => ref({}) })

const emit = defineEmits<{ (e: 'updateCategories', value: SlotVizCategories[]): void }>()

const selectedCategories = useCookie<SlotVizCategories[]>(
  COOKIE_KEY.SLOT_VIZ_SELECTED_CATEGORIES,
  {
    default: () => [
      'attestation',
      'proposal',
      'slashing',
      'sync',
      'visible',
      'initial',
    ],
  },
)

watch(() => persistedSlotVizCategories.value[dashboardKey.value],
  () => {
    if (persistedSlotVizCategories.value[dashboardKey.value]) {
      emit('updateCategories', persistedSlotVizCategories.value[dashboardKey.value])
    }
  },
)

const icons: MultiBarItem[] = [
  {
    component: IconSlotBlockProposal,
    tooltip: $t('slot_viz.filter.proposal'),
    value: 'proposal',
  },
  {
    component: IconSlotAttestation,
    tooltip: $t('slot_viz.filter.attestation'),
    value: 'attestation',
  },
  {
    component: IconSlotSync,
    tooltip: $t('slot_viz.filter.sync'),
    value: 'sync',
  },
  {
    component: IconSlotSlashing,
    tooltip: $t('slot_viz.filter.slashing'),
    value: 'slashing',
  },
  {
    className: 'visible-icon',
    icon: faEye,
    tooltip: $t('slot_viz.filter.visible'),
    value: 'visible',
  },
]

watch(
  () => props,
  () => {
    if (props.initiallyHideVisible !== undefined) {
      const initialIndex = selectedCategories.value.indexOf('initial')
      if (initialIndex < 0) {
        return
      }
      const categories = selectedCategories.value
      if (props.initiallyHideVisible) {
        categories.splice(initialIndex, 1)
        const visibleIndex = categories.indexOf('visible')
        if (visibleIndex >= 0) {
          categories.splice(visibleIndex, 1)
        }
      }
      else {
        categories.splice(initialIndex, 1, 'visible')
      }

      selectedCategories.value = categories
    }
  },
  { immediate: true },
)
</script>

<template>
  <BcToggleMultiBar
    v-model="selectedCategories"
    :buttons="icons"
  />
</template>

<style lang="scss" scoped>
@use "~/assets/css/main.scss";
 .dashboard-slot-viz-duty-toggle-loading-skeleton {
   @include main.container;
  height: 46px;
  width: 196px;
 }

:deep(.visible-icon) {
  margin-left: 4px;
  overflow: visible;
  position: relative;
}

:deep(.visible-icon):before {
  content: " ";
  background-color: var(--container-border-color);
  height: 100%;
  width: 1px;
  position: absolute;
  left: -5px;
}
</style>
