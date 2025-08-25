<script setup lang="ts">
// import { useStorage } from '@vueuse/core'
import type { MultiBarItem } from '~/types/multiBar'
import type { SlotVizCategories } from '~/types/dashboard/slotViz'

// type SlotVizCategoriesStorage = {
//   [dashboardId: string]: SlotVizCategories[],
// }

const { t: $t } = useTranslation()
// const {
// key,
// totalValidators,
// variant,
// isLargeDashboard,
// } = useDashboard()
// const validatorDashboardOverviewStore = useValidatorDashboardOverviewStore()
// const {
//   isLargeDashboard,
//   overview,
// } = storeToRefs(validatorDashboardOverviewStore)

// const persistedSelectedCategories = useStorage<SlotVizCategoriesStorage>('bc-dashboard-slot-viz-visibile-categories', {})

// const emit = defineEmits<{ (e: 'updateCategories', value: SlotVizCategories[]): void }>()

// const storageDashboardKey = computed(() => {
//   return key.value || 'guest-dashboard'
// })
// const selectedCategories = computed(() => {
//   const categories: SlotVizCategories[] = [
//     'attestation',
//     'proposal',
//     'slashing',
//     'sync',
//   ]

//   if ((variant.value !== 'shared-dashboard') || !isLargeDashboard.value) categories.push('visible')

//   return categories
// })
const icons: MultiBarItem[] = [
  {
    icon: 'cube',
    tooltip: $t('slot_viz.filter.proposal'),
    value: 'proposal',
  },
  {
    icon: 'file-signature',
    tooltip: $t('slot_viz.filter.attestation'),
    value: 'attestation',
  },
  {
    icon: 'sync',
    tooltip: $t('slot_viz.filter.sync'),
    value: 'sync',
  },
  {
    icon: 'user-slash',
    tooltip: $t('slot_viz.filter.slashing'),
    value: 'slashing',
  },
  {
    className: 'visible-icon',
    icon: 'eye',
    tooltip: $t('slot_viz.filter.visible'),
    value: 'visible',
  },
]

// onMounted(() => {
//   if (!persistedSelectedCategories.value[storageDashboardKey.value]) {
//     persistedSelectedCategories.value[storageDashboardKey.value] = selectedCategories.value
//   }
// })

// watch(() => overview.value, () => {
//   if (!persistedSelectedCategories.value[storageDashboardKey.value]) {
//     persistedSelectedCategories.value[storageDashboardKey.value] = selectedCategories.value
//   }
// })
// watch(() => persistedSelectedCategories.value[storageDashboardKey.value],
//   () => {
//     if (persistedSelectedCategories.value[storageDashboardKey.value]) {
//       emit('updateCategories', persistedSelectedCategories.value[storageDashboardKey.value])
//     }
//   },
//   { immediate: true },
// )

const modelValue = defineModel<SlotVizCategories[]>({
  required: true,
})
</script>

<template>
  <div>
    <!-- <div
      class="dashboard-slot-viz-duty-toggle-loading-skeleton"
    >
      <div class="dashboard-slot-viz-duty-toggle-loading-skeleton-content" />
    </div> -->
    <!-- <ClientOnly v-else> -->
    <BcToggleMultiBar
      v-model="modelValue"
      :buttons="icons"
    />
    <!-- </ClientOnly> -->
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

.dashboard-slot-viz-duty-toggle-loading-skeleton {
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
