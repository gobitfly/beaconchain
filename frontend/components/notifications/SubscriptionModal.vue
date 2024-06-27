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
  const changements = true
  dialogRef?.value.close(changements)
}
</script>

<template>
  <div v-if="newDataReceived" class="content">
    <div class="title">
      {{ t('notifications.subscriptions.dialog_title') }}
    </div>

    <div v-if="props?.validatorSub">
      <div class="explanation">
        {{ t(tPath+'explanation') }}
      </div>
      <NotificationsSubscriptionRow v-model="validatorSubModifiable.offlineValidator" :t-path="tPath+'offlineValidator'" :lacks-premium-subscription="false" />
      <NotificationsSubscriptionRow v-model="validatorSubModifiable.offlineGroup" :t-path="tPath+'offlineGroup'" :lacks-premium-subscription="!props.premiumUser" />
      <NotificationsSubscriptionRow v-model="validatorSubModifiable.missedAttestations" :t-path="tPath+'missedAttestations'" :lacks-premium-subscription="false" />
      <NotificationsSubscriptionRow v-model="validatorSubModifiable.proposedBlock" :t-path="tPath+'proposedBlock'" :lacks-premium-subscription="false" />
      <NotificationsSubscriptionRow v-model="validatorSubModifiable.upcomingProposal" :t-path="tPath+'upcomingProposal'" :lacks-premium-subscription="false" />
      <NotificationsSubscriptionRow v-model="validatorSubModifiable.syncCommittee" :t-path="tPath+'syncCommittee'" :lacks-premium-subscription="false" />
      <NotificationsSubscriptionRow v-model="validatorSubModifiable.withdrawn" :t-path="tPath+'withdrawn'" :lacks-premium-subscription="false" />
      <NotificationsSubscriptionRow v-model="validatorSubModifiable.shlashed" :t-path="tPath+'shlashed'" :lacks-premium-subscription="false" />
      <NotificationsSubscriptionRow v-model="validatorSubModifiable.realTime" :t-path="tPath+'realTime'" :lacks-premium-subscription="!props.premiumUser" />
    </div>

    <div v-else-if="props?.accountSub">
      <div class="explanation">
        {{ t(tPath+'explanation') }}
      </div>
      <NotificationsSubscriptionRow v-model="accountSubModifiable.incoming" :t-path="tPath+'incoming'" :lacks-premium-subscription="false" />
      <NotificationsSubscriptionRow v-model="accountSubModifiable.outgoing" :t-path="tPath+'outgoing'" :lacks-premium-subscription="false" />
      <NotificationsSubscriptionRow v-model="accountSubModifiable.erc20" :t-path="tPath+'erc20'" :lacks-premium-subscription="false" />
      <NotificationsSubscriptionRow v-model="accountSubModifiable.erc721" :t-path="tPath+'erc721'" :lacks-premium-subscription="false" />
      <NotificationsSubscriptionRow v-model="accountSubModifiable.erc1155" :t-path="tPath+'erc1155'" :lacks-premium-subscription="false" />
      <NotificationsSubscriptionRow v-model="accountSubModifiable.networks" :t-path="tPath+'networks'" :lacks-premium-subscription="false" />
      <NotificationsSubscriptionRow v-model="accountSubModifiable.ignoreSpam" :t-path="tPath+'ignoreSpam'" :lacks-premium-subscription="false" />
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
    @include fonts.subtitle_text;
    color: var(--primary-color);
    margin-bottom: var(--padding-small);
  }

  .explanation {
    color: var(--text-color-disabled);
  }

  .footer {
    display: flex;
    justify-content: right;
    margin-top: var(--padding);
    gap: var(--padding);
  }
}
</style>
