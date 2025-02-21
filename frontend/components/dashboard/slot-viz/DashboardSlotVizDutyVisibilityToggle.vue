<script setup lang="ts">
import { faEye } from '@fortawesome/pro-solid-svg-icons'
import { useStorage } from '@vueuse/core'
import type { MultiBarItem } from '~/types/multiBar'
import {
  IconSlotAttestation,
  IconSlotBlockProposal,
  IconSlotSlashing,
  IconSlotSync,
} from '#components'
import type { SlotVizCategories } from '~/types/dashboard/slotViz'
import type { VDBOverviewData } from '~/types/api/validator_dashboard'

type SlotVizCategoriesStorage = {
  [dashboardId: string]: SlotVizCategories[],
}

const {
  overviewData,
} = defineProps<{
  overviewData?: VDBOverviewData,
}>()

const { t: $t } = useTranslation()
const {
  dashboardKey, hasGuestDashboardKeyChanged, isSharedDashboard,
} = useDashboardKey()
const validatorDashboardStore = useValidatorDashboardStore()
const {
  isLargeDashboard,
} = storeToRefs(validatorDashboardStore)

const persistedSelectedCategories = useStorage<SlotVizCategoriesStorage>('bc-dashboard-slot-viz-visibile-categories', {})

const emit = defineEmits<{ (e: 'updateCategories', value: SlotVizCategories[]): void }>()

const storageDashboardKey = computed(() => {
  return dashboardKey.value || 'empty-guest-dashboard'
})
const selectedCategories = computed(() => {
  const categories: SlotVizCategories[] = [
    'attestation',
    'proposal',
    'slashing',
    'sync',
  ]

  if (!isSharedDashboard.value || !isLargeDashboard.value) categories.push('visible')

  return categories
})
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

onMounted(() => {
  if (!persistedSelectedCategories.value[storageDashboardKey.value]) {
    persistedSelectedCategories.value[storageDashboardKey.value] = selectedCategories.value
  }
})

watch(() => overviewData, () => {
  if (!persistedSelectedCategories.value[storageDashboardKey.value]) {
    persistedSelectedCategories.value[storageDashboardKey.value] = selectedCategories.value
  }
})
watch(() => persistedSelectedCategories.value[storageDashboardKey.value],
  () => {
    if (persistedSelectedCategories.value[storageDashboardKey.value]) {
      emit('updateCategories', persistedSelectedCategories.value[storageDashboardKey.value])
    }
  },
  { immediate: true },
)
watch(() => dashboardKey.value, (_, prevDashboardKey) => {
  // Whenever a guest dashboard key changes, we remove the old value from the storage in
  // order to avoid edge cases where a dashboard changes back to a previou key,
  // and to avoid accumulating unused dashboard keys in the storage.
  if (hasGuestDashboardKeyChanged) {
    const {
      [prevDashboardKey]: _, ...otherSavedDashboardKeys
    } = persistedSelectedCategories.value

    persistedSelectedCategories.value = otherSavedDashboardKeys
  }
})
</script>

<template>
  <div>
    <ClientOnly>
      <BcToggleMultiBar
        v-if="persistedSelectedCategories[storageDashboardKey]"
        v-model="persistedSelectedCategories[storageDashboardKey]"
        :buttons="icons"
      />
      <div
        v-else
        class="dashboard-slot-viz-duty-toggle-placeholder"
      >
        <div class="dashboard-slot-viz-duty-toggle-placeholder-content" />
      </div>

      <template #fallback>
        <div
          class="dashboard-slot-viz-duty-toggle-placeholder"
        >
          <div class="dashboard-slot-viz-duty-toggle-placeholder-content" />
        </div>
      </template>
    </ClientOnly>
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/main.scss";

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

.dashboard-slot-viz-duty-toggle-placeholder {
   @include main.container;
  height: 46px;
  width: 196px;
  padding: 7px 10px;
  animation: pulse 1.2s infinite;

  &-content {
    width: 100%;
    height: 100%;
    background-color: var(--container-border-color);
    border-radius: var(--border-radius);
  }

  // this keyframe imitates Tailwind's 'pulse' animation until we install it.
  // https://tailwindcss.com/docs/animation
  @keyframes pulse {
    50% {
      opacity: 0.5;
    }
  }
}
</style>
