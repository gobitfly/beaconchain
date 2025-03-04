<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faArrowUpRightFromSquare } from '@fortawesome/pro-solid-svg-icons'
import { useValidatorDashboardOverviewStore } from '~/stores/dashboard/useValidatorDashboardOverviewStore'
import { DashboardValidatorSubsetModal } from '#components'

const { t: $t } = useTranslation()

const validatorDashoboardOverviewStore = useValidatorDashboardOverviewStore()
const { overview } = storeToRefs(validatorDashoboardOverviewStore)

const validatorsOffline = computed(() => overview.value?.validators.offline ?? 0)
const validatorsOnline = computed(() => overview.value?.validators.online ?? 0)

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

const efficiencyInfos = [
  {
    label: $t('statistics.last_24h'),
    value: formatPercent(overview.value?.efficiency.last_24h ?? 0, {
      isFraction: false,
      maximumFractionDigits: 2,
    }),
  },
  {
    label: $t('statistics.last_7d'),
    value: formatPercent(overview.value?.efficiency.last_7d ?? 0, {
      isFraction: false,
      maximumFractionDigits: 2,
    }),
  },
  {
    label: $t('statistics.last_30d'),
    value: formatPercent(overview.value?.efficiency.last_30d ?? 0, {
      isFraction: false,
      maximumFractionDigits: 2,
    }),
  },
  {
    label: $t('statistics.all_time'),
    value: formatPercent(overview.value?.efficiency.all_time ?? 0, {
      isFraction: false,
      maximumFractionDigits: 2,
    }),
  },
]

const apr = computed(
  () => formatPercent((overview.value?.apr.last_30d.el ?? 0) + (overview.value?.apr.last_30d.cl ?? 0), {
    isFraction: false,
    maximumFractionDigits: 2,
  }),
)

const getText = (value: undefined | { cl: number, el: number }) => `${formatPercent(value?.cl ?? 0, {
  isFraction: false,
  maximumFractionDigits: 2,
})} (CL) ${formatPercent(value?.el ?? 0, {
  isFraction: false,
  maximumFractionDigits: 2,
})} (EL)`

const aprInfos = [
  {
    label: $t('statistics.last_24h'),
    value: getText(overview.value?.apr.last_24h),
  },
  {
    label: $t('statistics.last_7d'),
    value: getText(overview.value?.apr.last_7d),
  },
  {
    label: $t('statistics.last_30d'),
    value: getText(overview.value?.apr.last_30d),
  },
  {
    label: $t('statistics.all_time'),
    value: getText(overview.value?.apr.all_time),
  },
]
</script>

<template>
  <div class="container">
    <DashboardValidatorOverviewItem
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
        <BcFormatAmount
          :value="overview?.balances.total ?? '0'"
        />
      </template>
      <template #tooltip>
        <div>
          <section>
            <span>
              {{ $t('dashboard.validator.overview.validators_balance.balance_total_tooltip') }}:
            </span>
            <BcFormatAmount
              :value="overview?.balances.total ?? '0'"
            />
          </section>
          <section>
            <span>
              {{ $t('dashboard.validator.overview.validators_balance.balance_effective') }}:
            </span>
            <BcFormatAmount
              :value="overview?.balances.effective ?? '0'"
            />
          </section>
          <section>
            <span>
              {{ $t('dashboard.validator.overview.validators_balance.balance_staked') }}:
            </span>
            <BcFormatAmount
              :value="overview?.balances.staked_eth ?? '0'"
            />
          </section>
        </div>
      </template>
    </DashboardValidatorOverviewItem>
    <DashboardValidatorOverviewItem
      :infos="efficiencyInfos"
      :title="$t('dashboard.validator.overview.24h_efficiency')"
    >
      {{ formatPercent(overview?.efficiency.last_24h ?? 0, {
        isFraction: false,
        maximumFractionDigits: 2,
      }) }}
    </DashboardValidatorOverviewItem>
    <DashboardValidatorOverviewItem
      :title="$t('dashboard.validator.overview.30d_rewards')"
    >
      <BcFormatAmount
        :currency-items="[{
          consensusLayerValue: overview?.rewards.last_30d.cl ?? '0',
          executionLayerValue: overview?.rewards.last_30d.el ?? '0',
        }]"
        has-sign-display
        has-tooltip
      />
      <template #tooltip>
        <div class="tooltip">
          <section>
            <h3 class="tooltip-title">
              {{ $t('statistics.last_24h') }}
            </h3>
            <div>
              CL: <BcFormatAmount
                :value="overview?.rewards.last_24h.cl ?? '0'"
                target-currency="clDisplayCurrency"
                has-additional-selected-currency-main
                has-higher-precision
              />
            </div>
            <div>
              EL: <BcFormatAmount
                :value="overview?.rewards.last_24h.el ?? '0'"
                source-currency="elCurrency"
                target-currency="elDisplayCurrency"
                has-additional-selected-currency-main
                has-higher-precision
              />
            </div>
          </section>
          <section>
            <h3 class="tooltip-title">
              {{ $t('statistics.last_7d') }}
            </h3>
            <div>
              CL: <BcFormatAmount
                :value="overview?.rewards.last_7d.cl ?? '0'"
                target-currency="clDisplayCurrency"
                has-additional-selected-currency-main
                has-higher-precision
              />
            </div>
            <div>
              EL: <BcFormatAmount
                :value="overview?.rewards.last_7d.el ?? '0'"
                source-currency="elCurrency"
                target-currency="elDisplayCurrency"
                has-additional-selected-currency-main
                has-higher-precision
              />
            </div>
          </section>
          <section>
            <h3 class="tooltip-title">
              {{ $t('statistics.last_30d') }}
            </h3>
            <div>
              CL: <BcFormatAmount
                :value="overview?.rewards.last_30d.cl ?? '0'"
                target-currency="clDisplayCurrency"
                has-additional-selected-currency-main
                has-higher-precision
              />
            </div>
            <div>
              EL: <BcFormatAmount
                :value="overview?.rewards.last_30d.el ?? '0'"
                source-currency="elCurrency"
                target-currency="elDisplayCurrency"
                has-additional-selected-currency-main
                has-higher-precision
              />
            </div>
          </section>
          <section>
            <h3 class="tooltip-title">
              {{ $t('statistics.all_time') }}
            </h3>
            <div>
              CL: <BcFormatAmount
                :value="overview?.rewards.all_time.cl ?? '0'"
                target-currency="clDisplayCurrency"
                has-additional-selected-currency-main
                has-higher-precision
              />
            </div>
            <div>
              EL: <BcFormatAmount
                :value="overview?.rewards.all_time.el ?? '0'"
                source-currency="elCurrency"
                target-currency="elDisplayCurrency"
                has-additional-selected-currency-main
                has-higher-precision
              />
            </div>
          </section>
        </div>
      </template>
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
.tooltip {
  min-width: 6.25rem;
}
.tooltip-title {
  text-align: center;
}
</style>
