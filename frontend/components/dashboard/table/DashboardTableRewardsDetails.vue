<script setup lang="ts">
import { DashboardValidatorEpochDutiesModal } from '#components'
import type { Icon } from '~/components/bc/icon/BcIcon.vue'
import type { VDBRewardsTableRow } from '~/types/api/validator_dashboard'

interface Props {
  groupName?: string,
  row: VDBRewardsTableRow,
}
const props = defineProps<Props>()

const { dashboardKey } = useDashboardKey()

const { t: $t } = useTranslation()
const { details } = useValidatorDashboardRewardsDetailsStore(
  dashboardKey.value,
  props.row.group_id,
  props.row.epoch,
)

const dialog = useDialog()

const {
  addCurrencies,
  elCurrency,
} = useCurrency()
const proposerTotal = computed(() => {
  return addCurrencies({
    currencyItems: [
      {
        value: details.value?.proposal_cl_att_inc_reward ?? 0,
      },
      {
        value: details.value?.proposal_cl_sync_inc_reward ?? 0,
      },
      {
        value: details.value?.proposal_cl_slashing_inc_reward ?? 0,
      },
      {
        sourceCurrency: elCurrency,
        value: details.value?.proposal_el_reward ?? 0,
      },
    ],
  })
})

const data = computed(() => {
  if (!details.value) {
    return
  }

  const proposer = [
    {
      consensusLayerValue: details.value.proposal_cl_att_inc_reward,
      label: $t('dashboard.validator.rewards.proposer_rewards_cl_att'),
    },
    {
      consensusLayerValue: details.value.proposal_cl_sync_inc_reward,
      label: $t('dashboard.validator.rewards.proposer_rewards_cl_sync'),
    },
    {
      consensusLayerValue: details.value.proposal_cl_slashing_inc_reward,
      label: $t('dashboard.validator.rewards.proposer_rewards_cl_slash'),
    },
    {
      executionLayerValue: details.value.proposal_el_reward,
      label: $t('dashboard.validator.rewards.proposer_rewards_el'),
    },
  ]

  const rewards = [
    {
      icon: 'attestaion-source' as Icon,
      label: $t('dashboard.validator.rewards.attestation_source'),
      value: details.value.attestations_source,
    },
    {
      icon: 'attestaion-target' as Icon,
      label: $t('dashboard.validator.rewards.attestation_target'),
      value: details.value.attestations_target,
    },
    {
      icon: 'attestaion-head' as Icon,
      label: $t('dashboard.validator.rewards.attestation_head'),
      value: details.value.attestations_head,
    },
    {
      icon: 'cube' as Icon,
      label: $t('dashboard.validator.rewards.block'),
      value: {
        income: proposerTotal.value,
        status_count: details.value.proposal_status_count,
      },
    },
    {
      icon: 'sync' as Icon,
      label: $t('dashboard.validator.rewards.sync'),
      tooltip: formatMultiPartSpan(
        $t,
        'dashboard.validator.rewards.tooltip.sync',
        [ 'no-wrap' ],
      ),
      value: details.value.sync,
    },
    {
      icon: 'user-slash' as Icon,
      label: $t('dashboard.validator.rewards.slashing'),
      tooltip: formatMultiPartSpan(
        $t,
        'dashboard.validator.rewards.tooltip.slashing',
        [
          'slash-after no-wrap',
          ' no-wrap',
        ],
      ),
      value: details.value.slashing,
    },
    {
      icon: 'snooze' as Icon,
      label: $t('dashboard.validator.rewards.inactivity'),
      value: details.value.inactivity,
    },
  ].map((reward) => {
    const hasNoReward = !reward?.value?.status_count?.failed && !reward?.value?.status_count?.success
    const className = hasNoReward ? 'text-disabled' : ''
    return {
      ...reward,
      className,
      hasNoReward,
    }
  })
  return {
    proposer,
    rewards,
  }
})

const openDuties = () => {
  dialog.open(DashboardValidatorEpochDutiesModal, {
    data: {
      dashboardKey: dashboardKey.value,
      epoch: props.row.epoch,
      groupId: props.row.group_id,
      groupName: props.groupName,
    },
  })
}
</script>

<template>
  <div class="background">
    <div
      v-if="details"
      class="details-container"
    >
      <div>
        <div class="small-screen-value">
          <b><BcTableAgeHeader class="label" /></b>
          <div class="value">
            <BcTableDateTime :unix-timestamp="row.epoch" />
          </div>
        </div>
        <div class="small-screen-value">
          <div class="label">
            <b>{{ $t("dashboard.validator.col.duty") }}</b>
          </div>
          <div class="value">
            <DashboardTableValueDuty
              :duty="row.duty"
              class="detail-duty"
            />
          </div>
        </div>
      </div>
      <div class="rewards-container">
        <div class="rewards-group">
          <div class="col icon">
            <div
              v-for="item in data?.rewards"
              :key="item.label"
              class="row"
              :class="item.className"
            >
              <BcIcon
                v-if="item.icon"
                :name="item.icon"
              />
            </div>
            <div
              class="row"
            >
              <BcIcon
                name="sigma"
              />
            </div>
          </div>
          <div class="col label">
            <div
              v-for="item in data?.rewards"
              :key="item.label"
              class="label"
              :class="item.className"
            >
              {{ item.label }}
            </div>
            <div
              class="label"
            >
              {{ $t('dashboard.validator.rewards.total') }}
            </div>
          </div>
          <div class="col count">
            <BcTooltip
              v-for="item in data?.rewards"
              :key="item.label"
              :text="item.tooltip"
              class="row"
              :render-text-as-html="true"
              tooltip-class="text-align-left"
            >
              <DashboardTableEfficiency
                v-if="!item.hasNoReward"
                :success="item.value?.status_count?.success!"
                :failed="item.value?.status_count?.failed!"
                :absolute="true"
              />
              <div
                v-else
                class="text-disabled"
              >
                0 / 0
              </div>
            </BcTooltip>
            <div>
              <BcButtonIcon
                screenreader-text="dashboard.validator.rewards.open_duties_details"
                class="link popout"
                name="arrow-upright-from-square"
                @click="openDuties"
              />
            </div>
          </div>
          <div class="col value">
            <BcFormatAmount
              v-for="item in data?.rewards"
              :key="item.label"
              :value="item.value.income"
              :class="item.className"
              has-color
              has-sign-display
              has-tooltip
            />
            <div>
              <BcFormatAmount
                :currency-items="[{
                  executionLayerValue: props.row.reward.el,
                  consensusLayerValue: props.row.reward.cl,
                }]"
                has-color
                has-sign-display
                has-tooltip
              />
            </div>
          </div>
        </div>
        <div class="proposer-group">
          <div
            v-for="item in data?.proposer"
            :key="item.label"
            class="row"
            :class="{
              'text-disabled': !(item.executionLayerValue && item.consensusLayerValue),
            }"
          >
            <div class="label">
              {{ item.label }}
            </div>
            <BcFormatAmount
              :currency-items="[{
                consensusLayerValue: item.consensusLayerValue ?? '0',
                executionLayerValue: item.executionLayerValue ?? '0',
              }]"
              has-color
              has-tooltip
              has-sign-display
            />
          </div>
          <div
            class="row"
            :class="{
              'text-disabled': proposerTotal === '0',
            }"
          >
            <div class="label">
              {{ $t('dashboard.validator.rewards.proposer_rewards_total') }}
            </div>
            <BcFormatAmount
              :value="proposerTotal"
              has-color
              has-tooltip
              has-sign-display
            />
          </div>
        </div>
      </div>
    </div>
    <div v-else>
      <BcLoadingSpinner
        class="spinner"
        :loading="true"
        alignment="center"
      />
    </div>
  </div>
</template>

<style lang="scss" scoped>
.background {
  color: var(--container-color);
  background-color: var(--container-background);
}

.spinner {
  padding: var(--padding-large);
}

.details-container {
  font-size: var(--small_text_font_size);
  padding: 14px 28px;

  .small-screen-value {
    display: none;
    margin-bottom: var(--padding-large);
    width: 360px;
    justify-content: space-between;

    .label {
      width: 90px;
    }

    .value {
      flex-grow: 1;
      text-align: right;

      :deep(.detail-duty) {
        justify-content: flex-end;

        .group {
          &:nth-child(2) {
            &:after {
              content: unset;
            }
          }
        }
      }
    }
  }

  .rewards-container {
    display: flex;
    flex-wrap: wrap;
    gap: var(--padding-xl);
    font-size: var(--small_text_font_size);

    .rewards-group {
      display: flex;
      width: 360px;

      .col {
        > div,
        > span {
          height: 32px;
          padding: var(--padding-small);
          text-wrap: nowrap;

          &:last-child {
            border-top: solid 1px var(--container-border-color);
            font-weight: var(--small_text_bold_font_weight);
          }
        }

        &.icon {
          svg {
            height: 14px;
            width: 18px;
          }
        }

        &.count {
          display: flex;
          flex-direction: column;
        }

        &.value {
          display: flex;
          flex-direction: column;
          flex-grow: 1;
          align-items: flex-end;

          > div {
            width: 100%;
            text-align: end;
          }
        }
      }
    }

    .proposer-group {
      width: 360px;

      .row {
        height: 32px;
        padding: var(--padding-small) 0;
        display: flex;
        justify-content: space-between;
        width: 330px;

        &:last-child {
          border-top: solid 1px var(--container-border-color);
          font-weight: var(--small_text_bold_font_weight);
        }
      }
    }
  }
}

@media screen and (max-width: 1180px) {
  .details-container {
    .small-screen-value {
      display: flex;
    }
  }
}

@media screen and (max-width: 900px) {
  .details-container {
    width: 400px;
    padding: var(--padding) var(--padding-large);

    .rewards-container {
      flex-direction: column-reverse;
      gap: var(--padding-large);
      width: 100%;

      .rewards-group {
        width: 100%;
      }

      .proposer-group {
        width: 100%;

        .row {
          width: 100%;
        }
      }
    }
  }
}

@media screen and (max-width: 420px) {
  .details-container {
    width: 100%;
  }

  .details-container {
    .small-screen-value {
      width: 100%;
    }
  }
}
</style>
