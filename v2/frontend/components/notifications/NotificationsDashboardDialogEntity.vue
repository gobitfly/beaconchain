<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import {
  faAlarmSnooze,
  faArrowsRotate,
  faChartLineUp,
  faCube,
  faFileSignature,
  faGlobe,
  faMoneyBill,
  faPowerOff,
  faRocket,
  faUserSlash,
} from '@fortawesome/pro-solid-svg-icons'
import type { NotificationDashboardsTableRow } from '~/types/api/notifications'

const { t: $t } = useTranslation()

const {
  props,
} = useBcDialog<{ identifier: string } & Pick<NotificationDashboardsTableRow, 'dashboard_id' | 'epoch' | 'group_id' | 'group_name'>>()

const store = useNotificationsDashboardDetailsStore()

const search = ref('')
const {
  data: details,
  status,
} = useAsyncData(
  'notifications-dashboard-details',
  () => store.getDetails({
    dashboard_id: props.value?.dashboard_id ?? 0,
    epoch: props.value?.epoch ?? 0,
    group_id: props.value?.group_id ?? 0,
    search: search.value.length ? search.value : undefined,
  }).then(response => response.data),
  {
    watch: [ search ],
  })
defineEmits<{ (e: 'filter-changed', value: string): void }>()
const { converter } = useValue()
const formatValueWei = (value: string) => {
  return converter.value.weiToValue(`${value}`, { fixedDecimalCount: 5 })
    .label
}
</script>

<template>
  <div class="notifications-dashboard-dialog-entity">
    <header class="notifications-dashboard-dialog-entity__header">
      <h2>
        <BcText
          variant="lg"
        >
          {{ $t('notifications.dashboards.dialog.entity.title') }}
        </BcText>
        <BcText
          variant="md"
          is-dimmed
        >
          ({{ $t('common.epoch') }} {{ props?.epoch }})
        </BcText>
      </h2>
      <div class="notifications-dashboard-dialog-entity__subheader">
        <h3>
          {{ props?.group_name }} ({{ details?.dashboard_name }})
        </h3>
        <BcContentFilter
          :search-placeholder="$t('common.index')"
          :is-loading="status=== 'pending'"
          @filter-changed="search = $event"
        />
      </div>
    </header>
    <main
      class="notifications-dashboard-dialog-entity__content"
    >
      <BcAccordion
        v-if="details?.validator_offline?.length"
        :items="details?.validator_offline"
        :info-copy="$t('notifications.dashboards.dialog.entity.validator_offline')"
      >
        <template #headingIcon>
          <FontAwesomeIcon :icon="faPowerOff" class="notifications-dashboard-dialog-entity__icon__red" />
        </template>
        <template #heading>
          {{ $t('notifications.dashboards.dialog.entity.validator_offline') }} ({{ details?.validator_offline?.length ?? 0 }})
        </template>
        <template #item="{ item: validatorIndex }">
          <BcLink
            :to="`/validator/${validatorIndex}`"
            class="link"
          >
            {{ validatorIndex }}
          </BcLink>
        </template>
      </BcAccordion>
      <BcAccordion
        v-if="details?.proposal_missed?.length"
        :items="details?.proposal_missed"
        :info-copy="$t('notifications.dashboards.dialog.entity.proposal_missed')"
      >
        <template #headingIcon>
          <FontAwesomeIcon
            :icon="faCube"
            class="notifications-dashboard-dialog-entity__icon__red"
          />
        </template>
        <template #heading>
          {{ $t('notifications.dashboards.dialog.entity.proposal_missed') }} ({{ details?.proposal_missed?.length ?? 0 }})
        </template>
        <template #item="{ item: proposal }">
          <BcLink
            class="link"
            :to="`/validator/${proposal.index}`"
          >
            {{ proposal.index }}
          </BcLink>
          <template v-if="proposal.slots.length">
            [<BcLink
              v-for="block in proposal.slots"
              :key="block"
              :to="`/block/${block}`"
              class="notifications-dashboard-dialog-entity__list-item link"
            >
              {{ block }}
            </BcLink>]
          </template>
        </template>
      </BcAccordion>
      <BcAccordion
        v-if="details?.proposal_success?.length"
        :items="details?.proposal_success"
        :info-copy="$t('notifications.dashboards.dialog.entity.proposal_done')"
      >
        <template #headingIcon>
          <FontAwesomeIcon
            :icon="faCube"
            class="notifications-dashboard-dialog-entity__icon__green"
          />
        </template>
        <template #heading>
          {{ $t('notifications.dashboards.dialog.entity.proposal_done') }} ({{ details?.proposal_success?.length ?? 0 }})
        </template>
        <template #item="{ item: proposalDone }">
          <BcLink
            :to="`/validator/${proposalDone.index}`"
            class="link"
          >
            {{ proposalDone.index }}
          </BcLink>
          <template v-if="proposalDone.blocks.length">
            [<BcLink
              v-for="block in proposalDone.blocks"
              :key="block"
              :to="`/block/${block}`"
              class="notifications-dashboard-dialog-entity__list-item link"
            >
              {{ block }}
            </BcLink>]
          </template>
        </template>
      </BcAccordion>
      <BcAccordion
        v-if="details?.slashed?.length"
        :items="details?.slashed"
        :info-copy="$t('notifications.dashboards.dialog.entity.slashed')"
      >
        <template #headingIcon>
          <FontAwesomeIcon :icon="faUserSlash" class="notifications-dashboard-dialog-entity__icon__red" />
        </template>
        <template #heading>
          {{ $t('notifications.dashboards.dialog.entity.slashed') }} ({{ details?.slashed?.length ?? 0 }})
        </template>
        <template #item="{ item: slashedValidatorIndex }">
          <BcLink
            :to="`/validator/${slashedValidatorIndex}`"
            class="link"
          >
            {{ slashedValidatorIndex }}
          </BcLink>
        </template>
      </BcAccordion>
      <BcAccordion
        v-if="details?.sync?.length"
        :items="details?.sync"
        :info-copy="$t('notifications.dashboards.dialog.entity.sync_committee')"
      >
        <template #headingIcon>
          <FontAwesomeIcon :icon="faArrowsRotate" class="notifications-dashboard-dialog-entity__icon__green" />
        </template>
        <template #heading>
          {{ $t('notifications.dashboards.dialog.entity.sync_committee') }} ({{ details?.sync?.length ?? 0 }})
        </template>
        <template #item="{ item: syncCommitteIndex }">
          <BcLink
            :to="`/validator/${syncCommitteIndex}`"
            class="link"
          >
            {{ syncCommitteIndex }}
          </BcLink>
        </template>
      </BcAccordion>
      <BcAccordion
        v-if="details?.attestation_missed?.length"
        :items="details?.attestation_missed"
        :info-copy="$t('notifications.dashboards.dialog.entity.attestation_missed')"
      >
        <template #heading>
          {{ $t('notifications.dashboards.dialog.entity.attestation_missed') }} ({{ details?.attestation_missed?.length ?? 0 }})
        </template>
        <template #headingIcon>
          <FontAwesomeIcon
            :icon="faFileSignature"
            class="notifications-dashboard-dialog-entity__icon__red"
          />
        </template>
        <template #item="{ item: attestation }">
          <BcLink
            :to="`/validator/${attestation.index}`"
            class="link"
          >
            {{ attestation.index }}
          </BcLink>
          (<BcLink
            :to="`/epoch/${attestation.epoch}`"
            class="link"
          >
            {{ $t('common.epoch') }}
            {{ attestation.epoch }}
          </BcLink>)
        </template>
      </BcAccordion>
      <BcAccordion
        v-if="details?.withdrawal?.length"
        :items="details?.withdrawal"
        :info-copy="$t('notifications.dashboards.dialog.entity.withdrawal')"
      >
        <template #heading>
          {{ $t('notifications.dashboards.dialog.entity.withdrawal') }} ({{ details?.withdrawal?.length ?? 0 }})
        </template>
        <template #headingIcon>
          <FontAwesomeIcon
            :icon="faMoneyBill"
            class="notifications-dashboard-dialog-entity__icon__green"
          />
        </template>
        <template #item="{ item: withdrawalItem }">
          <BcLink
            :to="`/validator/${withdrawalItem.index}`"
            class="link"
          >
            {{ withdrawalItem.index }}
          </BcLink>
          ({{ formatValueWei(withdrawalItem.amount) }})
        </template>
      </BcAccordion>
      <BcAccordion
        v-if="details?.validator_online?.length"
        :items="details?.validator_online"
        :info-copy="$t('notifications.dashboards.dialog.entity.validator_back_online')"
      >
        <template #headingIcon>
          <FontAwesomeIcon
            :icon="faGlobe"
            class="notifications-dashboard-dialog-entity__icon__green"
          />
        </template>
        <template #heading>
          {{ $t('notifications.dashboards.dialog.entity.validator_back_online') }} ({{ details?.validator_online?.length ?? 0 }})
        </template>
        <template #item="{ item: validator }">
          <BcLink
            :to="`/validator/{{ validator.index }}`"
            class="link"
          >
            {{ validator.index }}
          </BcLink>
          ({{ validator.epoch_count }} {{ $t('common.epoch', validator.epoch_count) }})<!--
            this will remove white space in html
          -->
        </template>
      </BcAccordion>
      <BcAccordion
        v-if="details?.group_efficiency_below"
        :item="details?.group_efficiency_below"
      >
        <template #headingIcon>
          <FontAwesomeIcon
            :icon="faChartLineUp"
            class="notifications-dashboard-dialog-entity__icon__red"
          />
        </template>
        <template #heading>
          {{ $t('notifications.dashboards.dialog.entity.group_efficiency') }}
        </template>
        <template #item="{ item: groupEfficiencyBelow }">
          {{ details?.group_name }}
          (<BcLink :to="`/dashboard/${props?.dashboard_id}`">
            {{ details?.dashboard_name }}
          </BcLink>)
          {{ $t('notifications.dashboards.dialog.entity.group_efficiency_text', {
            percentage: formatFractionToPercent(groupEfficiencyBelow),
          }) }}
        </template>
      </BcAccordion>
      <BcAccordion
        v-if="details?.validator_offline_reminder?.length"
        :items="details?.validator_offline_reminder"
        :info-copy="$t('notifications.dashboards.dialog.entity.validator_offline_reminder')"
      >
        <template #headingIcon>
          <FontAwesomeIcon
            :icon="faAlarmSnooze"
            class="notifications-dashboard-dialog-entity__icon__red"
          />
        </template>
        <template #heading>
          {{ $t('notifications.dashboards.dialog.entity.validator_offline_reminder') }} ({{ details?.validator_offline_reminder?.length ?? 0 }})
        </template>
        <template #item="{ item: validatorOfflineReminderIndex }">
          <BcLink
            :to="`/validator/${validatorOfflineReminderIndex}`"
            class="link"
          >
            {{ validatorOfflineReminderIndex }}
          </BcLink>
        </template>
      </BcAccordion>
      <BcAccordion
        v-if="details?.proposal_upcoming?.length"
        :items="details?.proposal_upcoming"
        :info-copy="$t('notifications.dashboards.dialog.entity.upcoming_proposal')"
      >
        <template #headingIcon>
          <FontAwesomeIcon
            :icon="faCube"
            class="notifications-dashboard-dialog-entity__icon__green"
          />
        </template>
        <template #heading>
          {{ $t('notifications.dashboards.dialog.entity.upcoming_proposal') }} ({{ details?.proposal_upcoming?.length ?? 0 }})
        </template>
        <template #item="{ item: upcomingProposal }">
          <BcLink
            class="link"
            :to="`/validator/${upcomingProposal.index}`"
          >
            {{ upcomingProposal.index }}
          </BcLink>
          <template v-if="upcomingProposal.slots.length">
            [<BcLink
              v-for="block in upcomingProposal.slots"
              :key="block"
              :to="`/block/${block}`"
              class="notifications-dashboard-dialog-entity__list-item link"
            >
              {{ block }}
            </BcLink>]
          </template>
        </template>
      </BcAccordion>
      <BcAccordion
        v-if="details?.min_collateral?.length"
        :items="details?.min_collateral"
        :info-copy="$t('notifications.dashboards.dialog.entity.min_collateral')"
      >
        <template #headingIcon>
          <FontAwesomeIcon
            :icon="faRocket"
            class="notifications-dashboard-dialog-entity__icon__red"
          />
        </template>
        <template #heading>
          {{ $t('notifications.dashboards.dialog.entity.min_collateral') }} ({{ details?.min_collateral?.length ?? 0 }})
        </template>
        <template #item="{ item: minCollateral }">
          <BcFormatHash
            :ens="minCollateral.ens"
            :hash="minCollateral.hash"
            type="public_key"
            no-copy
            class="overwrite-block"
          >
            {{ minCollateral }}
          </BcFormatHash>
        </template>
      </BcAccordion>
      <BcAccordion
        v-if="details?.max_collateral?.length"
        :items="details?.max_collateral"
        :info-copy="$t('notifications.dashboards.dialog.entity.max_collateral')"
      >
        <template #headingIcon>
          <FontAwesomeIcon
            :icon="faRocket"
            class="notifications-dashboard-dialog-entity__icon__green"
          />
        </template>
        <template #heading>
          {{ $t('notifications.dashboards.dialog.entity.max_collateral') }} ({{ details?.max_collateral?.length ?? 0 }})
        </template>
        <template #item="{ item: maxCollateral }">
          <BcFormatHash
            :ens="maxCollateral.ens"
            :hash="maxCollateral.hash"
            type="public_key"
            no-copy
            class="overwrite-block"
          >
            {{ maxCollateral }}
          </BcFormatHash>
        </template>
      </BcAccordion>
    </main>
  </div>
</template>

<style scoped lang="scss">
@use '~/assets/css/breakpoints' as *;

:deep(div.format-hash.overwrite-block) {
  display: inline-flex
}
.notifications-dashboard-dialog-entity {
  @media (min-width: $breakpoint-md) {
    min-width: 42rem;
  }
}
.notifications-dashboard-dialog-entity__header{
  display: flex;
  flex-direction: column;
  gap: 0.938rem;
}
.notifications-dashboard-dialog-entity__subheader {
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
}
.notifications-dashboard-dialog-entity__content {
  margin-top: 1.25rem;
  display:flex;
  flex-direction: column;
  gap: 0.625rem;
}
.notifications-dashboard-dialog-entity__icon__green {
  color: #7DC382;
}
.notifications-dashboard-dialog-entity__icon__red {
  color: #F3454A;
}
.notifications-dashboard-dialog-entity__list-item:not(:last-child)::after {
content: ', ';
}
</style>
