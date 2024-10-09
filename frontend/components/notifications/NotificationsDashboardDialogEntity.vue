<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faGlobe } from '@fortawesome/pro-solid-svg-icons'
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

const validatorsBackOnline = computed(() => {
  return details.value?.data?.validator_back_online
})
const groupsBackOnline = computed(() => {
  return details.value?.data?.group_back_online
})
const proposalMissed = computed(() => {
  return details.value?.data?.proposal_missed
})
const proposalDone = computed(() => {
  return details.value?.data?.proposal_done
})
const slashed = computed(() => {
  return details.value?.data?.slashed
})
const syncCommittee = computed(() => {
  return details.value?.data?.sync_committee
})(
const attestaitonMissed = computed(() => {
  return details.value?.data?.attestation_missed
})
const withdrawl = computed(() => {
  
})
</script>

<template>
  <pre>
    {{ props }}
  </pre>
  <div class="notifications-dashboard-dialog-entity">
    <header>
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
      <BcAccordion
        :items="groupsBackOnline"
      >
        <template #heading>
          {{ $t('notifications.dashboards.dialog.entity.group_back_online', [groupsBackOnline?.length ?? 0]) }}
        </template>
        <template #headingIcon>
          <FontAwesomeIcon :icon="faGlobe" />
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
        :items="validatorsBackOnline"
      >
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
        :items="proposalMissed"
      >
        <template #heading>
          {{ $t('notifications.dashboards.dialog.entity.proposal_missed', [proposalMissed?.length ?? 0]) }}
        </template>
        <template #item="{ item: proposal }">
          <BcLink
            :to="`/validator/{{ validator.index }}`"
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
        :items="proposalDone"
      >
        <template #heading>
          {{ $t('notifications.dashboards.dialog.entity.proposal_done', [proposalDone?.length ?? 0]) }}
        </template>
        <template #item="{ item: propossal }">
          <BcLink
            to=""
          >
            {{ propossal.index }}
          </BcLink>
          {{ propossal.blocks }}<!--
            this will remove white space in html
          -->
        </template>
      </BcAccordion>
      <BcAccordion
        :items="slashed"
      >
        <template #heading>
          {{ $t('notifications.dashboards.dialog.entity.slashed', [slashed?.length ?? 0]) }}
        </template>
        <template #item="{ item: slashed }">
          <BcLink
            to=""
          >
            <pre>{{ slashed }}</pre>
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
      <BcAccordion
        :items="proposalMissed"
      >
        <template #heading>
          {{ $t('notifications.dashboards.dialog.entity.proposal_done', [proposalDone?.length ?? 0]) }}
        </template>
        <template #item="{ item: propossal }">
          <BcLink
            to=""
          >
            {{ propossal.index }}
          </BcLink>
          {{ propossal.blocks }}<!--
            this will remove white space in html
          -->
        </template>
      </BcAccordion>
      <BcAccordion
        :items="attestaitonMissed"
      >
        <template #heading>
          {{ $t('notifications.dashboards.dialog.entity.proposal_done', [attestaitonMissed?.length ?? 0]) }}
        </template>
        <template #item="{ item: attestaiton }">
          <BcLink
            to=""
          >
            {{ attestaiton.index }}
          </BcLink>
          {{ attestaiton.blocks }}<!--
            this will remove white space in html
          -->
        </template>
      </BcAccordion>
      <BcAccordion
        :items=""
      />
    </main>
  </div>
</template>

<style scoped lang="scss">
.notifications-dashboard-dialog-entity {
  width: 44rem;
}
.notifications-dashboard-dialog-entity__content {
  margin-top: 1.25rem;
}
</style>
