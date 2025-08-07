<script setup lang="ts">
import type { NotificationDashboardsTableRow } from '~/types/api/notifications'
import BcIcon from '~/components/bc/icon/BcIcon.vue'

const { t: $t } = useTranslation()

const {
  props,
} = useBcDialog<Pick<NotificationDashboardsTableRow, 'dashboard_id' | 'epoch' | 'group_id' | 'group_name'> & { identifier: string }>()

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
const v1Domain = useV1Domain()
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
          <BcIcon
            name="power-off"
            class="notifications-dashboard-dialog-entity__icon__red"
          />
        </template>
        <template #heading>
          {{ $t('notifications.dashboards.dialog.entity.validator_offline') }} ({{ details?.validator_offline?.length ?? 0 }})
        </template>
        <template #item="{ item: validatorIndex }">
          <BcLink
            :to="`${v1Domain}/validator/${validatorIndex}`"
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
          <BcIcon
            name="cube"
            class="notifications-dashboard-dialog-entity__icon__red"
          />
        </template>
        <template #heading>
          {{ $t('notifications.dashboards.dialog.entity.proposal_missed') }} ({{ details?.proposal_missed?.length ?? 0 }})
        </template>
        <template #item="{ item: proposal }">
          <BcLink
            class="link"
            :to="`${v1Domain}/validator/${proposal.index}`"
          >
            {{ proposal.index }}
          </BcLink>
          <template v-if="proposal.slots.length">
            [<BcLink
              v-for="block in proposal.slots"
              :key="block"
              :to="`${v1Domain}/block/${block}`"
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
          <BcIcon
            name="cube"
            class="notifications-dashboard-dialog-entity__icon__green"
          />
        </template>
        <template #heading>
          {{ $t('notifications.dashboards.dialog.entity.proposal_done') }} ({{ details?.proposal_success?.length ?? 0 }})
        </template>
        <template #item="{ item: proposalDone }">
          <BcLink
            :to="`${v1Domain}/validator/${proposalDone.index}`"
            class="link"
          >
            {{ proposalDone.index }}
          </BcLink>
          <template v-if="proposalDone.blocks.length">
            [<BcLink
              v-for="block in proposalDone.blocks"
              :key="block"
              :to="`${v1Domain}/block/${block}`"
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
          <BcIcon
            name="user-slash"
            class="notifications-dashboard-dialog-entity__icon__red"
          />
        </template>
        <template #heading>
          {{ $t('notifications.dashboards.dialog.entity.slashed') }} ({{ details?.slashed?.length ?? 0 }})
        </template>
        <template #item="{ item: slashedValidatorIndex }">
          <BcLink
            :to="`${v1Domain}/validator/${slashedValidatorIndex}`"
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
          <BcIcon
            name="rotate"
            class="notifications-dashboard-dialog-entity__icon__green"
          />
        </template>
        <template #heading>
          {{ $t('notifications.dashboards.dialog.entity.sync_committee') }} ({{ details?.sync?.length ?? 0 }})
        </template>
        <template #item="{ item: syncCommitteIndex }">
          <BcLink
            :to="`${v1Domain}/validator/${syncCommitteIndex}`"
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
          <BcIcon
            name="file-signature"
            class="notifications-dashboard-dialog-entity__icon__red"
          />
        </template>
        <template #item="{ item: attestation }">
          <BcLink
            :to="`${v1Domain}/validator/${attestation.index}`"
            class="link"
          >
            {{ attestation.index }}
          </BcLink>
          (<BcLink
            :to="`${v1Domain}/epoch/${attestation.epoch}`"
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
          <BcIcon
            name="money-bill"
            class="notifications-dashboard-dialog-entity__icon__green"
          />
        </template>
        <template #item="{ item: withdrawalItem }">
          <BcLink
            :to="`${v1Domain}/validator/${withdrawalItem.index}`"
            class="link"
          >
            {{ withdrawalItem.index }}
          </BcLink>
          <BcFormatAmount
            v-slot="{ value }"
            :value="withdrawalItem.amount"
          >
            ({{ value }})
          </BcFormatAmount>
        </template>
      </BcAccordion>
      <BcAccordion
        v-if="details?.validator_online?.length"
        :items="details?.validator_online"
        :info-copy="$t('notifications.dashboards.dialog.entity.validator_back_online')"
      >
        <template #headingIcon>
          <BcIcon
            name="globe"
            class="notifications-dashboard-dialog-entity__icon__green"
          />
        </template>
        <template #heading>
          {{ $t('notifications.dashboards.dialog.entity.validator_back_online') }} ({{ details?.validator_online?.length ?? 0 }})
        </template>
        <template #item="{ item: validator }">
          <BcLink
            :to="`${v1Domain}/validator/{{ validator.index }}`"
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
          <BcIcon
            name="chart-line-up"
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
            percentage: formatPercent(groupEfficiencyBelow, {
              maximumFractionDigits: 0,
            }),
          }) }}
        </template>
      </BcAccordion>
      <BcAccordion
        v-if="details?.validator_offline_reminder?.length"
        :items="details?.validator_offline_reminder"
        :info-copy="$t('notifications.dashboards.dialog.entity.validator_offline_reminder')"
      >
        <template #headingIcon>
          <BcIcon
            name="alarm-snooze"
            class="notifications-dashboard-dialog-entity__icon__red"
          />
        </template>
        <template #heading>
          {{ $t('notifications.dashboards.dialog.entity.validator_offline_reminder') }} ({{ details?.validator_offline_reminder?.length ?? 0 }})
        </template>
        <template #item="{ item: validatorOfflineReminderIndex }">
          <BcLink
            :to="`${v1Domain}/validator/${validatorOfflineReminderIndex}`"
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
          <BcIcon
            name="cube"
            class="notifications-dashboard-dialog-entity__icon__green"
          />
        </template>
        <template #heading>
          {{ $t('notifications.dashboards.dialog.entity.upcoming_proposal') }} ({{ details?.proposal_upcoming?.length ?? 0 }})
        </template>
        <template #item="{ item: upcomingProposal }">
          <BcLink
            class="link"
            :to="`${v1Domain}/validator/${upcomingProposal.index}`"
          >
            {{ upcomingProposal.index }}
          </BcLink>
          <template v-if="upcomingProposal.slots.length">
            [<BcLink
              v-for="block in upcomingProposal.slots"
              :key="block"
              :to="`${v1Domain}/block/${block}`"
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
          <BcIcon
            name="rocket"
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
          <BcIcon
            name="rocket"
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
