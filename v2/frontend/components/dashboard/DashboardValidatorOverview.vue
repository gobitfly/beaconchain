<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faArrowUpRightFromSquare } from '@fortawesome/pro-solid-svg-icons'
import { useValidatorDashboardOverviewStore } from '~/stores/dashboard/useValidatorDashboardOverviewStore'
import type { ClElValue } from '~/types/api/common'
import {
  type NumberOrString, TimeFrames,
} from '~/types/value'
import { totalElClNumbers } from '~/utils/bigMath'
import { DashboardValidatorSubsetModal } from '#components'

const { t: $t } = useTranslation()
const { converter } = useValue()

const { overview } = useValidatorDashboardOverviewStore()

const formatValueWei = (value: NumberOrString): NumberOrString => {
  return converter.value.weiToValue(`${value}`, { fixedDecimalCount: 4 })
    .label
}

const createInfo = (
  key: string,
  value: ClElValue<number | string>,
  formatFunction: (value: Partial<NumberOrString>) => NumberOrString,
) => {
  const clValue = formatFunction(value.cl)
  const elValue = formatFunction(value.el)
  return {
    label: $t(`statistics.${key}`),
    value: `${clValue} (CL) ${elValue} (EL)`,
  }
}

const rewardFull = computed(() => converter.value.weiToValue(
  totalElCl(overview.value?.rewards.last_30d ?? {
    cl: '0',
    el: '0',
  }), { addPlus: true }).fullLabel)
const reward = computed(() => converter.value.weiToValue(
  totalElCl(overview.value?.rewards.last_30d ?? {
    cl: '0',
    el: '0',
  }), { addPlus: true }).label)

const validatorsOffline = computed(() => overview.value?.validators.offline ?? 0)
const validatorsOnline = computed(() => overview.value?.validators.online ?? 0)
const validatorsInfos = computed(() =>
  [
    {
      label: $t('dashboard.validator.overview.validators_balance.balance_staked'),
      value: formatValueWei(overview.value?.balances.staked_eth ?? 0),
    },
    {
      label: $t('dashboard.validator.overview.validators_balance.balance_effective'),
      value: formatValueWei(overview.value?.balances.effective ?? 0),
    },
  ],
)

const dialog = useDialog()
const { dashboardKey } = useDashboardKey()
const { getDashboardLabel } = useUserDashboardStore()
const openValidatorModal = () => {
  dialog.open(DashboardValidatorSubsetModal, {
    data: {
      context: 'dashboard',
      dashboardKey: dashboardKey.value,
      dashboardName: getDashboardLabel(dashboardKey.value, 'validator'),
      timeFrame: 'last_24h',
    },
  })
}

const efficiencyInfos = computed(() =>
  TimeFrames.map(k => ({
    label: $t(`statistics.${k}`),
    value: formatPercent(overview.value?.efficiency[k] ?? 0),
  })),
)

const rewardsInfos = TimeFrames.map(k =>
  createInfo(k, overview.value?.rewards[k] ?? {
    cl: '0',
    el: '0',
  }, formatValueWei),
)

const apr = computed(() => formatPercent(totalElClNumbers(overview.value?.apr.last_30d ?? {
  cl: 0,
  el: 0,
})),
)

const aprInfos = TimeFrames.map(k =>
  createInfo(k, overview.value?.apr[k] ?? {
    cl: 0,
    el: 0,
  }, formatToPercent),
)
</script>

<template>
  <div class="container">
    <DashboardValidatorOverviewItem
      :infos="validatorsInfos"
      :title="$t('dashboard.validator.overview.online_validators')"
    >
      <span :class="{ positive: validatorsOnline }">
        {{ validatorsOnline }}
      </span> |
      <span :class="{ negative: validatorsOffline }">
        {{ validatorsOffline }}
      </span>
      <BcButtonIcon
        :sr-text="$t('dashboard.validator.overview.open_validator_overview_modal')"
        @click="openValidatorModal"
      >
        <FontAwesomeIcon
          class="link optical-correction"
          :icon="faArrowUpRightFromSquare"
        />
      </BcButtonIcon>
      <template #additionalInfo>
        {{ $t('dashboard.validator.overview.validators_balance.balance_total') }}
        <span class="bold">
          {{ formatValueWei(overview?.balances.total ?? 0) }}
        </span>
      </template>
    </DashboardValidatorOverviewItem>
    <DashboardValidatorOverviewItem
      :infos="efficiencyInfos"
      :title="$t('dashboard.validator.overview.24h_efficiency')"
    >
      {{ formatToPercent(overview?.efficiency.last_24h ?? 0) }}
    </DashboardValidatorOverviewItem>
    <DashboardValidatorOverviewItem
      :infos="rewardsInfos"
      :title="$t('dashboard.validator.overview.30d_rewards')"
    >
      <BcTooltip
        :text="rewardFull"
        :fit-content="true"
      >
        {{ reward }}
      </BcTooltip>
    </DashboardValidatorOverviewItem>
    <DashboardValidatorOverviewItem
      :infos="aprInfos"
      :title="$t('dashboard.validator.overview.30d_apr')"
    >
      {{ apr }}
    </DashboardValidatorOverviewItem>
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/main.scss";
.container {
  @include main.container;
  display: flex;
  align-items: center;
  justify-content: space-between;
  overflow-x: auto;
  margin-top: var(--padding-large);
  transform: translateY(0px); // hack: on safari top-border is not shown
  gap: 50px;
  height: 101px;
  padding-left: var(--padding-xl);
  padding-right: var(--padding-xl);
}

.optical-correction {
  transform: translateY(1px);
}
</style>
