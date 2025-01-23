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

const validatorDashoboardOverviewStore = useValidatorDashboardOverviewStore()
const { overview } = storeToRefs(validatorDashoboardOverviewStore)
const {
  addCurrencies,
  clCurrency,
  displayCurrencyDefault,
  elCurrency,
  formatAmount,
  selectedCurrencyMain,
} = useCurrency()

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

const rewardLast30d = computed(() => {
  return addCurrencies({
    currencyItems: [
      {
        sourceCurrency: clCurrency,
        value: overview.value?.rewards.last_30d.cl ?? '0',
      },
      {
        sourceCurrency: elCurrency,
        value: overview.value?.rewards.last_30d.el ?? '0',
      },
    ],
  })
})

const validatorsOffline = computed(() => overview.value?.validators.offline ?? 0)
const validatorsOnline = computed(() => overview.value?.validators.online ?? 0)
const validatorsInfos = computed(() =>
  [
    {
      label: $t('dashboard.validator.overview.validators_balance.balance_total_tooltip'),
      value: `${formatAmount(overview.value?.balances.total ?? 0)}`,
    },
    {
      label: $t('dashboard.validator.overview.validators_balance.balance_effective'),
      value: `${formatAmount(overview.value?.balances.effective ?? 0)}`,
    },
    {
      label: $t('dashboard.validator.overview.validators_balance.balance_staked'),
      value: `${formatAmount(overview.value?.balances.staked_eth ?? 0)}`,
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
  TimeFrames.map(timeFrame => ({
    label: $t(`statistics.${timeFrame}`),
    value: formatToPercent(overview.value?.efficiency[timeFrame] ?? 0),
  })),
)

const format = (value: string, type: 'consensusLayer' | 'executionLayer') => {
  const sourceCurrency = type === 'consensusLayer' ? clCurrency : elCurrency
  const targetCurrency = type === 'consensusLayer' ? displayCurrencyDefault.consensusLayer : displayCurrencyDefault.executionLayer
  const amount = formatAmount(value, {
    sourceCurrency,
    targetCurrency,
  })
  const isDifferentCurrency = targetCurrency !== selectedCurrencyMain.value
  if (isDifferentCurrency) {
    const amountRelativeToSelectedCurrency = formatAmount(value, {
      sourceCurrency,
      targetCurrency: selectedCurrencyMain.value,
    })
    return `${amount} (${amountRelativeToSelectedCurrency})`
  }
  return `${amount}`
}
const rewardsInfos = computed(() => [
  {
    label: $t('statistics.last_24h'),
    value: `CL: ${
      format(overview.value?.rewards.last_24h.cl ?? '0', 'consensusLayer')
    } EL: ${
      format(overview.value?.rewards.last_24h.el ?? '0', 'executionLayer')
    }`,
  },
  {
    label: $t('statistics.last_7d'),
    value: `CL: ${
      format(overview.value?.rewards.last_7d.cl ?? '0', 'consensusLayer')
    } EL: ${
      format(overview.value?.rewards.last_7d.el ?? '0', 'executionLayer')
    }`,
  },
  {
    label: $t('statistics.last_30d'),
    value: `CL: ${
      format(overview.value?.rewards.last_30d.cl ?? '0', 'consensusLayer')
    } EL: ${
      format(overview.value?.rewards.last_30d.el ?? '0', 'executionLayer')
    }`,
  },
  {
    label: $t('statistics.all_time'),
    value: `CL: ${
      format(overview.value?.rewards.all_time.cl ?? '0', 'consensusLayer')
    } EL: ${
      format(overview.value?.rewards.all_time.el ?? '0', 'executionLayer')
    }`,
  },
])

const apr = computed(() => formatToPercent(totalElClNumbers(overview.value?.apr.last_30d ?? {
  cl: 0,
  el: 0,
})))
const aprInfos = TimeFrames.map(timeFrame =>
  createInfo(timeFrame, overview.value?.apr[timeFrame] ?? {
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
        :screenreader-text="$t('dashboard.validator.overview.open_validator_overview_modal')"
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
          {{ formatAmount(overview?.balances.total ?? '0') }}
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
        :text="formatAmount(rewardLast30d, {
          signDisplay: 'always',
          hasHigherPrecision: true,
        })"
        :fit-content="true"
      >
        {{ formatAmount(rewardLast30d, {
          signDisplay: 'always',
        }) }}
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
