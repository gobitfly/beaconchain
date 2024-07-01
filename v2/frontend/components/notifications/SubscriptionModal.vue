<script setup lang="ts">
import type { ValidatorSubscriptionState, AccountSubscriptionState } from '~/types/subscriptionModal'

type ValidatorSubscriptionStateComplete = ValidatorSubscriptionState & {
  offlineGroup: number,
  realTime: boolean
}

interface Props {
  validatorSub?: ValidatorSubscriptionState,
  accountSub?: AccountSubscriptionState,
  premiumUser: boolean
}

const { props, dialogRef } = useBcDialog<Props>({ showHeader: false })
const { t } = useI18n()

const newDataReceived = ref<number>(0) // used in <template> to trigger Vue to refresh or hide the content
let tPath: string
const validatorSubModifiable = ref({} as ValidatorSubscriptionStateComplete)
const accountSubModifiable = ref({} as AccountSubscriptionState)
const all = ref(false)

watch(props, (props) => {
  if (!props || (!props.validatorSub && !props.accountSub)) {
    newDataReceived.value = 0
    return
  }
  newDataReceived.value++
  if (props.validatorSub) {
    tPath = 'notifications.subscriptions.validators.'
    validatorSubModifiable.value = { offlineGroup: -1, realTime: false, ...structuredClone(toRaw(props.validatorSub)) }
  } else {
    tPath = 'notifications.subscriptions.accounts.'
    accountSubModifiable.value = structuredClone(toRaw(props.accountSub!))
  }
}, { immediate: true })

const closeDialog = () => {
  const modified = true
  dialogRef?.value.close(modified)
}
</script>

<template>
  <div v-if="newDataReceived" class="content">
    <div class="title">
      {{ t('notifications.subscriptions.dialog_title') }}
    </div>

    <div v-if="t(tPath+'explanation')" class="explanation">
      {{ t(tPath+'explanation') }}
    </div>

    <div v-if="props?.validatorSub">
      <NotificationsSubscriptionRow v-model="validatorSubModifiable.offlineValidator" :t-path="tPath+'offline_validator'" :lacks-premium-subscription="false" class="row" />
      <NotificationsSubscriptionRow
        v-model="validatorSubModifiable.offlineGroup"
        :t-path="tPath+'offline_group'"
        :lacks-premium-subscription="!props.premiumUser"
        input-type="percent"
        :default="10"
        class="row"
      />
      <NotificationsSubscriptionRow v-model="validatorSubModifiable.missedAttestations" :t-path="tPath+'missed_attestations'" :lacks-premium-subscription="false" class="row" />
      <NotificationsSubscriptionRow v-model="validatorSubModifiable.proposedBlock" :t-path="tPath+'proposed_block'" :lacks-premium-subscription="false" class="row" />
      <NotificationsSubscriptionRow v-model="validatorSubModifiable.upcomingProposal" :t-path="tPath+'upcoming_proposal'" :lacks-premium-subscription="false" class="row" />
      <NotificationsSubscriptionRow v-model="validatorSubModifiable.syncCommittee" :t-path="tPath+'sync_committee'" :lacks-premium-subscription="false" class="row" />
      <NotificationsSubscriptionRow v-model="validatorSubModifiable.withdrawn" :t-path="tPath+'withdrawn'" :lacks-premium-subscription="false" class="row" />
      <NotificationsSubscriptionRow v-model="validatorSubModifiable.slashed" :t-path="tPath+'slashed'" :lacks-premium-subscription="false" class="row" />
      <NotificationsSubscriptionRow v-model="validatorSubModifiable.realTime" :t-path="tPath+'real_time'" :lacks-premium-subscription="!props.premiumUser" class="row" />
      <hr>
      <NotificationsSubscriptionRow v-model="all" :t-path="tPath+'all'" :lacks-premium-subscription="false" class="row" />
    </div>

    <div v-else-if="props?.accountSub">
      <NotificationsSubscriptionRow v-model="accountSubModifiable.incoming" :t-path="tPath+'incoming'" :lacks-premium-subscription="false" class="row" />
      <NotificationsSubscriptionRow v-model="accountSubModifiable.outgoing" :t-path="tPath+'outgoing'" :lacks-premium-subscription="false" class="row" />
      <NotificationsSubscriptionRow v-model="accountSubModifiable.erc20" :t-path="tPath+'erc20'" :lacks-premium-subscription="false" input-type="amount" class="row" />
      <NotificationsSubscriptionRow v-model="accountSubModifiable.erc721" :t-path="tPath+'erc721'" :lacks-premium-subscription="false" class="row" />
      <NotificationsSubscriptionRow v-model="accountSubModifiable.erc1155" :t-path="tPath+'erc1155'" :lacks-premium-subscription="false" class="row" />
      <hr>
      <NotificationsSubscriptionRow v-model="all" :t-path="tPath+'all'" :lacks-premium-subscription="false" class="row" />
      <NotificationsSubscriptionRow v-model="accountSubModifiable.networks" :t-path="tPath+'networks'" :lacks-premium-subscription="false" input-type="networks" class="row" />
      <NotificationsSubscriptionRow v-model="accountSubModifiable.ignoreSpam" :t-path="tPath+'ignore_spam'" :lacks-premium-subscription="false" class="row" />
    </div>

    <div class="footer">
      <Button type="button" :label="t('notifications.subscriptions.save')" @click="closeDialog" />
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use '~/assets/css/fonts.scss';

.content {
  display: flex;
  flex-direction: column;

  .title {
    @include fonts.dialog_header;
    text-align: center;
    margin-bottom: var(--padding-large);
  }

  .explanation {
    margin-bottom: var(--padding);
    @include fonts.small_text;
    opacity: 0.6;
  }

  .row {
    margin-top: 14px;
    margin-bottom: 14px;
  }

  .footer {
    display: flex;
    justify-content: right;
    margin-top: var(--padding);
    gap: var(--padding);
  }
}
</style>
