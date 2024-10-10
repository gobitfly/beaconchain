<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import {
  faAlarmSnooze,
  faArrowsRotate,
  faCube,
  faFileSignature,
  faGlobe,
  faMoneyBill,
  faPowerOff,
  faUserSlash,
} from '@fortawesome/pro-solid-svg-icons'
import type { NotificationDashboardsTableRow } from '~/types/api/notifications'

// const model = defineModel(false)
const { t: $t } = useTranslation()

const {
  // close,
  props,
} = useBcDialog<{ identifier: string } & Pick<NotificationDashboardsTableRow, 'dashboard_id' | 'epoch' | 'group_id' | 'group_name'>>()

const store = useNotificationsDashboardDetailsStore()

const { data: details } = useAsyncData(() => store.getDetails({
  // dashboard_id: props.value?.dashboard_id ?? 0,
  dashboard_id: 5426, // 💀 (personal dashboard id) remove after development
  epoch: props.value?.epoch ?? 0,
  group_id: props.value?.group_id ?? 0,
}))
watch(details, () => {
  // console.log(data)
  // store.addDetails({
  //   details: data.value,
  //   identifier: props.value?.identifier ?? '',
  // })
})

// 🥺
const validatorsOffline = computed(() => {
  return details.value?.data?.validator_offline
})
const groupsOffline = computed(() => {
  return details.value?.data?.group_offline
})
const validatorsBackOnline = computed(() => {
  return details.value?.data?.validator_back_online
})
const attestationsMissed = computed(() => {
  return details.value?.data?.attestation_missed
})
const groupsBackOnline = computed(() => {
  return details.value?.data?.group_back_online
})
const proposalsMissed = computed(() => {
  return details.value?.data?.proposal_missed
})
const proposalsDone = computed(() => {
  return details.value?.data?.proposal_done
})
const slashed = computed(() => {
  return details.value?.data?.slashed
})
const syncCommittee = computed(() => {
  return details.value?.data?.sync_committee
})
const withdrawls = computed(() => {
  return details.value?.data?.withdrawal
})
const validatorsOfflineReminder = computed(() => {
  return details.value?.data?.validator_offline_reminder
})
const groupsOfflineReminder = computed(() => {
  return details.value?.data?.group_offline_reminder
})
const upcomingProposals = computed(() => {
  return details.value?.data?.upcoming_proposals
})
// Todo: 🚨 Add it when BEDS-587 is merged.
// const max_collateral = computed(() => {
//   return details.value?.data?.max_collateral
// })
// const min_collateral = computed(() => {
//   return details.value?.data?.min_collateral
// })
</script>

<template>
  <!-- <pre>
    {{ details }}
  </pre> -->
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
      <h3>
        {{ props?.group_name }}
      </h3>
      <div id="notifications-management-search-placholder" />
    </header>
    <main
      v-if="details?.data"
      class="notifications-dashboard-dialog-entity__content"
    >
      <!-- <details v-for="(detailValue, detailKey) in details.data" :key="detailKey">
      <summary>
        {{ detailKey }}
      </summary>
      <p v-if="detailKey === 'validator_back_online'">
        <span v-for="(value, index) in detailKey" :key="`${value}-${key}`">
          {{ value }}
          <br>
          {{ key }}
        </span>
      </p>
    </details> -->
      <!-- validator offline -->
      <BcAccordion
        :items="validatorsOffline"
      >
        <template #headingIcon>
          <FontAwesomeIcon :icon="faPowerOff" class="notifications-dashboard-dialog-entity__icon__red" />
        </template>
        <template #heading>
          {{ $t('notifications.dashboards.dialog.entity.validatore_offline', [validatorsOffline?.length ?? 0]) }}
        </template>
        <template #item="{ item: validatorOffline }">
          <BcLink
            to=""
          >
            {{ validatorOffline }}
          </BcLink>
          <!--
            this will remove white space in html
          -->
        </template>
      </BcAccordion>
      <BcAccordion
        :items="groupsOffline"
      >
        <template #heading>
          {{ $t('notifications.dashboards.dialog.entity.group_offline', [groupsOffline?.length ?? 0]) }}
        </template>
        <template #headingIcon>
          <FontAwesomeIcon :icon="faPowerOff" class="notifications-dashboard-dialog-entity__icon__red" />
        </template>
        <template #item="{ item: groupOffline }">
          <span>
            {{ groupOffline.group_name }}
          </span>

          <BcLink
            :to="`/dashboard/${groupOffline.dashboard_id}`"
          >
            ({{ groupOffline.dashboard_id }})
          </BcLink>
        </template>
      </BcAccordion>
      <BcAccordion
        :items="proposalsMissed"
      >
        <template #headingIcon>
          <FontAwesomeIcon :icon="faCube" class="notifications-dashboard-dialog-entity__icon__red" />
        </template>
        <template #heading>
          {{ $t('notifications.dashboards.dialog.entity.proposal_missed', [proposalsMissed?.length ?? 0]) }}
        </template>
        <template #item="{ item: proposal }">
          <BcLink
            to=""
          >
            {{ proposal.index }}
          </BcLink>
          <!-- ({{ proposal.blocks }} {{ $t('common.epoch', validator.epoch_count) }}) -->
          <!--
            this will remove white space in html
          -->
          {{ proposal.blocks }}<!--
            this will remove white space in html
          -->
        </template>
      </BcAccordion>
      <BcAccordion
        :items="proposalsDone"
      >
        <template #headingIcon>
          <FontAwesomeIcon :icon="faCube" class="notifications-dashboard-dialog-entity__icon__green" />
        </template>
        <template #heading>
          {{ $t('notifications.dashboards.dialog.entity.proposal_done', [proposalsDone?.length ?? 0]) }}
        </template>
        <template #item="{ item: proposal }">
          <BcLink
            to=""
          >
            {{ proposal.index }}
          </BcLink>
          {{ proposal.blocks }}<!--
            this will remove white space in html
          -->
        </template>
      </BcAccordion>
      <BcAccordion
        :items="slashed"
      >
        <template #headingIcon>
          <FontAwesomeIcon :icon="faUserSlash" class="notifications-dashboard-dialog-entity__icon__red" />
        </template>
        <template #heading>
          {{ $t('notifications.dashboards.dialog.entity.slashed', [slashed?.length ?? 0]) }}
        </template>
        <template #item="{ item: slashed }">
          <BcLink
            to=""
          >
            {{ slashed }}
          </BcLink>
          {{ slashed }}<!--
            this will remove white space in html
          -->
        </template>
      </BcAccordion>
      <BcAccordion
        :items="syncCommittee"
      >
        <template #headingIcon>
          <FontAwesomeIcon :icon="faArrowsRotate" class="notifications-dashboard-dialog-entity__icon__green" />
        </template>
        <template #heading>
          {{ $t('notifications.dashboards.dialog.entity.sync_comittee', [syncCommittee?.length ?? 0]) }}
        </template>
        <template #item="{ item: sync_committee }">
          <BcLink
            to=""
          />
          {{ sync_committee }}<!--
            this will remove white space in html
          -->
        </template>
      </BcAccordion>
      <BcAccordion :items="attestationsMissed">
        <template #heading>
          {{ $t('notifications.dashboards.dialog.entity.attestation_missed', [attestationsMissed?.length ?? 0]) }}
        </template>
        <template #headingIcon>
          <FontAwesomeIcon :icon="faGlobe" class="notifications-dashboard-dialog-entity__icon__red" />
        </template>
        <template #item="{ item: attestationMissed }">
          <BcLink
            to=""
          >
            {{ attestationMissed.blocks }}
          </BcLink>
          <!--
            this will remove white space in html
          -->
        </template>
      </BcAccordion>
      <BcAccordion
        :items="withdrawls"
      >
        <template #headingIcon>
          <FontAwesomeIcon :icon="faMoneyBill" class="notifications-dashboard-dialog-entity__icon__green" />
        </template>
        <template #heading>
          {{ $t('notifications.dashboards.dialog.entity.withdrawl', [withdrawls?.length ?? 0]) }}
        </template>
        <template #item="{ item: withdrawl }">
          <BcLink
            to=""
          >
            {{ withdrawl.index }}
          </BcLink>
          {{ withdrawl.blocks }}<!-- this will remove white space in html
          -->
        </template>
      </BcAccordion>
      <BcAccordion
        :items="validatorsBackOnline"
      >
        <template #headingIcon>
          <FontAwesomeIcon :icon="faGlobe" class="notifications-dashboard-dialog-entity__icon__green" />
        </template>
        <template #heading>
          {{ $t('notifications.dashboards.dialog.entity.validator_back_online', [validatorsBackOnline?.length ?? 0]) }}
        </template>
        <template #item="{ item: validator }">
          <BcLink
            :to="`/validator/{{ validator.index }}`"
          >
            {{ validator.index }}
          </BcLink>
          ({{ validator.epoch_count }} {{ $t('common.epoch', validator.epoch_count) }})<!--
            this will remove white space in html
          -->
        </template>
      </BcAccordion>
      <BcAccordion
        :items="groupsBackOnline"
      >
        <template #heading>
          {{ $t('notifications.dashboards.dialog.entity.group_back_online', [groupsBackOnline?.length ?? 0]) }}
        </template>
        <template #headingIcon>
          <FontAwesomeIcon :icon="faGlobe" class="notifications-dashboard-dialog-entity__icon__green" />
        </template>
        <template #item="{ item: group }">
          <span>
            {{ group.group_name }}
          </span>
          <span>
            [{{ group.epoch_count }}&nbsp;{{ $t('common.epoch', group.epoch_count) }}]
          </span>
          <BcLink
            :to="`/dashboard/${group.dashboard_id}`"
          >
            <!-- Todo: 🚨 put in dashboard name here -->
            (Dashboard {{ group.dashboard_id }})
          </BcLink>
        </template>
      </BcAccordion>
      <BcAccordion
        :items="validatorsOfflineReminder"
      >
        <template #headingIcon>
          <FontAwesomeIcon :icon="faAlarmSnooze" class="notifications-dashboard-dialog-entity__icon__red" />
        </template>
        <template #heading>
          {{ $t('notifications.dashboards.dialog.entity.validator_offline_reminder', [validatorsOfflineReminder?.length ?? 0]) }}
        </template>
        <template #item="{ item: validatorOfflineReminder }">
          <BcLink
            to=""
          >
            {{ validatorOfflineReminder }}
          </BcLink>
        </template>
      </BcAccordion>
      <BcAccordion
        :items="groupsOfflineReminder"
      >
        <template #headingIcon>
          <FontAwesomeIcon :icon="faAlarmSnooze" class="notifications-dashboard-dialog-entity__icon__red" />
        </template>
        <template #heading>
          {{ $t('notifications.dashboards.dialog.entity.group_offline_reminder', [groupsOfflineReminder?.length ?? 0]) }}
        </template>
        <template #item="{ item: groupOfflineReminder }">
          <BcLink
            to=""
          >
            {{ groupOfflineReminder.dashboard_id }}
          </BcLink>
          {{ groupOfflineReminder.group_name }}<!-- this will remove white space in html
          -->
        </template>
      </BcAccordion>
      <BcAccordion
        :items="upcomingProposals"
      >
        <template #headingIcon>
          <FontAwesomeIcon :icon="faCube" class="notifications-dashboard-dialog-entity__icon__green" />
        </template>
        <template #heading>
          {{ $t('notifications.dashboards.dialog.entity.upcoming_proposal', [upcomingProposals?.length ?? 0]) }}
        </template>
        <template #item="{ item: upcomingProposal }">
          <BcLink
            to=""
          >
            {{ upcomingProposal }}
          </BcLink>
          ({{ upcomingProposal }})<!-- this will remove white space in html
          -->
        </template>
      </BcAccordion>
      <!-- 🚨 TODO: Min Collateral reached -->
      <!-- 🚨 TODO: Max Collateral reached -->
    </main>
  </div>
</template>

<style scoped lang="scss">
.notifications-dashboard-dialog-entity {
  width: 44rem;
  height:40.875rem;
}
.notifications-dashboard-dialog-entity__header{
  display: flex;
  flex-direction: column;
  gap: 0.938rem;
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
</style>
