<script setup lang="ts">
import { faArrowUpRightFromSquare, faSigma, faSnooze } from '@fortawesome/pro-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { IconSlotBlockProposal, IconSlotHeadAttestation, IconSlotSlashing, IconSlotSourceAttestation, IconSlotSync, IconSlotTargetAttestation } from '#components'
import type { VDBGroupRewardsDetails, VDBRewardsTableRow } from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'
import type BcTooltip from '~/components/bc/BcTooltip.vue'

interface Props {
  dashboardKey: DashboardKey
  row: VDBRewardsTableRow
}
const props = defineProps<Props>()

const { t: $t } = useI18n()
const { details } = useValidatorDashboardRewardsDetailsStore(props.dashboardKey, props.row.group_id, props.row.epoch)

const data = computed(() => {
  if (!details.value) {
    return
  }

  // TODO: check where to get the data, once the api structs have changed
  const proposer = [
    {
      label: $t('dashboard.validator.rewards.proposer_rewards_cl_att'),
      value: details.value.proposal_el_reward
    },
    {
      label: $t('dashboard.validator.rewards.proposer_rewards_cl_sync'),
      value: details.value.proposal_el_reward
    },
    {
      label: $t('dashboard.validator.rewards.proposer_rewards_cl_slash'),
      value: details.value.proposal_el_reward
    },
    {
      label: $t('dashboard.validator.rewards.proposer_rewards_el'),
      value: details.value.proposal_el_reward
    },
    {
      label: $t('dashboard.validator.rewards.proposer_rewards_total'),
      value: details.value.proposal_el_reward
    }
  ]

  const rewards = [
    {
      svg: IconSlotSourceAttestation,
      label: $t('dashboard.validator.rewards.attestation_source'),
      value: details.value.attestations_source
    },
    {
      svg: IconSlotTargetAttestation,
      label: $t('dashboard.validator.rewards.attestation_target'),
      value: details.value.attestations_target
    },
    {
      svg: IconSlotHeadAttestation,
      label: $t('dashboard.validator.rewards.attestation_head'),
      value: details.value.attestations_head
    },
    {
      svg: IconSlotBlockProposal,
      label: $t('dashboard.validator.rewards.block'),
      value: details.value.proposal
    },
    {
      svg: IconSlotSync,
      label: $t('dashboard.validator.rewards.sync'),
      value: details.value.sync,
      tooltip: formatMultiPartSpan($t, 'dashboard.validator.rewards.tooltip.sync', ['no-wrap'])
    },
    {
      svg: IconSlotSlashing,
      label: $t('dashboard.validator.rewards.slashing'),
      value: details.value.slashing,
      tooltip: formatMultiPartSpan($t, 'dashboard.validator.rewards.tooltip.slashing', ['slash-after no-wrap', ' no-wrap'])
    },
    {
      icon: faSnooze,
      label: $t('dashboard.validator.rewards.inactivity'),
      value: details.value.attestations_head // TODO: replace with inactivity once we get it
    },
    {
      icon: faSigma,
      label: $t('dashboard.validator.rewards.total'),
      value: {
        income: totalElCl(props.row.reward)
      } as Partial<VDBGroupRewardsDetails>,
      isTotal: true
    }
  ]
  return {
    proposer,
    rewards
  }
})

const openDuties = () => {
  // TODO: implement modal
  alert('open details')
}

</script>
<template>
  <div class="background">
    <div v-if="details" class="details-container">
      <div>
        <div class="small-screen-value">
          <div class="label">
            <b>{{ $t('common.age') }}</b>
          </div>
          <div class="value">
            <BcFormatTimePassed :value="row.epoch" />
          </div>
        </div>
        <div class="small-screen-value">
          <div class="label">
            <b>{{ $t('dashboard.validator.col.duty') }}</b>
          </div>
          <div class="value">
            <DashboardTableValueDuty :duty="row.duty" class="detail-duty" />
          </div>
        </div>
      </div>
      <div class="rewards-container">
        <div class="rewards-group">
          <div class="col icon">
            <div v-for="item in data?.rewards" :key="item.label" class="row">
              <component :is="item.svg" v-if="item.svg" />
              <FontAwesomeIcon v-if="item.icon" :icon="item.icon" />
            </div>
          </div>
          <div class="col label">
            <div v-for="item in data?.rewards" :key="item.label" class="label">
              {{ item.label }}
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
                v-if="item.value.status_count"
                :success="item.value.status_count.success"
                :failed="item.value.status_count.failed"
              />
              <div v-if="item.isTotal">
                <FontAwesomeIcon class="link popout" :icon="faArrowUpRightFromSquare" @click="openDuties" />
              </div>
            </BcTooltip>
          </div>
          <div class="col value">
            <BcFormatValue
              v-for="item in data?.rewards"
              :key="item.label"
              :value="item.value.income"
              :use-colors="true"
              :options="{ addPlus: true }"
            />
          </div>
        </div>
        <div class="proposer-group">
          <div v-for="item in data?.proposer" :key="item.label" class="row">
            <div class="label">
              {{ item.label }}
            </div>
            <BcFormatValue :value="item.value" :use-colors="true" :options="{ addPlus: true }" />
          </div>
        </div>
      </div>
    </div>
    <div v-else>
      <BcLoadingSpinner class="spinner" :loading="true" alignment="center" />
    </div>
  </div>
</template>

<style lang="scss" scoped>
.background {
  color: var(--container-color);
  background-color: var(--container-background);
}

.spinner{
  padding: var(--padding-large);
}

.details-container {
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

    .rewards-group {
      display: flex;
      width: 360px;

      .col {

        >div,
        >span {
          height: 32px;
          padding: var(--padding-small);
          text-wrap: nowrap;

          &:last-child {
            border-top: solid 1px var(--container-border-color);
            font-weight: var(--standard_text_medium_font_weight);
          }
        }

        &.icon{
          svg{
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

          >div {
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
          font-weight: var(--standard_text_medium_font_weight);
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
