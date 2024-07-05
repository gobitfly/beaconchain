<script setup lang="ts">
import type { ValidatorSubscriptionState, AccountSubscriptionState, CheckboxAndNumber } from '~/types/subscriptionModal'
import type { ChainID } from '~/types/network'

interface Props {
  validatorSub?: ValidatorSubscriptionState,
  accountSub?: AccountSubscriptionState,
  premiumUser: boolean
}

const MinimumTimeBetweenAPIcalls = 500 // ms
const DefaultValueOfValidatorOptionsNeedingPremium = {
  offlineGroup: -10, // means "10% and unchecked"
  realTime: false
  // ... add lines here if some options become premium in the future
}
const DefaultValueOfAccountOptionsNeedingPremium = {
  // ... add lines here if some options become premium in the future
}
const OptionsOutsideTheScopeOfCheckboxall: Array<keyof(ValidatorSubscriptionState & AccountSubscriptionState)> =
      ['networks', 'ignoreSpam']

type AllPossibleOptions = ValidatorSubscriptionState & AccountSubscriptionState & typeof DefaultValueOfValidatorOptionsNeedingPremium & typeof DefaultValueOfAccountOptionsNeedingPremium
type ModifiableOptions = Record<keyof AllPossibleOptions, CheckboxAndNumber|ChainID[]>

const { props, dialogRef } = useBcDialog<Props>({ showHeader: false })
const { t } = useI18n()

const tPath = ref('')
const modifiableOptions = ref({} as ModifiableOptions)
const allCheckbox = ref({ check: false } as CheckboxAndNumber)

const debouncer = useDebounceValue<ModifiableOptions>({} as ModifiableOptions, MinimumTimeBetweenAPIcalls)
watch(debouncer.value, sendUserPreferencesToAPI)

let newDataJustLoaded = true // is used by the watcher of `modifiableOptions` to know when it is unnecessary to ouptut apparent changes to the API

watch(props, (props) => {
  if (!props || (!props.validatorSub && !props.accountSub)) {
    return
  }
  newDataJustLoaded = true
  let options: AllPossibleOptions
  if (props.validatorSub) {
    tPath.value = 'notifications.subscriptions.validators.'
    options = { ...DefaultValueOfValidatorOptionsNeedingPremium, ...structuredClone(toRaw(props.validatorSub)) } as AllPossibleOptions
  } else {
    tPath.value = 'notifications.subscriptions.accounts.'
    options = { ...DefaultValueOfAccountOptionsNeedingPremium, ...structuredClone(toRaw(props.accountSub)) } as AllPossibleOptions
  }
  modifiableOptions.value = {} as ModifiableOptions
  for (const entry of Object.entries(options)) {
    const key = entry[0] as keyof typeof options
    if (Array.isArray(entry[1])) {
      modifiableOptions.value[key] = entry[1]
    } else {
      switch (typeof entry[1]) {
        case 'boolean' : modifiableOptions.value[key] = { check: entry[1], num: 0 }; break
        case 'number' : modifiableOptions.value[key] = { check: entry[1] >= 0, num: Math.abs(entry[1]) }; break
      }
    }
  }
}, { immediate: true })

watch(allCheckbox, (option) => {
  for (const k of Object.keys(modifiableOptions.value)) {
    const key = k as keyof ModifiableOptions
    if (isOptionAvailable(key) && !OptionsOutsideTheScopeOfCheckboxall.includes(key)) {
      (modifiableOptions.value[key] as CheckboxAndNumber).check = option.check
    }
  }
  // no need to call the API, the modifications that we did in `modifiableOptions.value` will trigger its watcher (that calls the API)
})

watch(modifiableOptions, (options) => {
  allCheckbox.value.check = true
  for (const k of Object.keys(modifiableOptions.value)) {
    const key = k as keyof ModifiableOptions
    if (isOptionAvailable(key) && !OptionsOutsideTheScopeOfCheckboxall.includes(key)) {
      allCheckbox.value.check &&= (modifiableOptions.value[key] as CheckboxAndNumber).check
    }
  }
  if (!newDataJustLoaded) {
    // it will call `sendUserPreferencesToAPI()`
    debouncer.bounce(options)
  }
  newDataJustLoaded = false
}, { immediate: true, deep: true })

function sendUserPreferencesToAPI () : void {
// envoyer ce qi achange
}

const isOptionAvailable = (key: string) => props.value?.premiumUser || !(key in DefaultValueOfValidatorOptionsNeedingPremium || key in DefaultValueOfAccountOptionsNeedingPremium)

const closeDialog = () => {
  const modified = true
  dialogRef?.value.close(modified)
}
</script>

<template>
  <div v-if="props && tPath" class="content">
    <div class="title">
      {{ t('notifications.subscriptions.dialog_title') }}
    </div>

    <div v-if="t(tPath+'explanation')" class="explanation">
      {{ t(tPath+'explanation') }}
    </div>

    <div v-if="props.validatorSub">
      <NotificationsSubscriptionRow v-model="modifiableOptions.offlineValidator" :t-path="tPath+'offline_validator'" :lacks-premium-subscription="!isOptionAvailable('offlineValidator')" class="row" />
      <NotificationsSubscriptionRow
        v-model="modifiableOptions.offlineGroup"
        :t-path="tPath+'offline_group'"
        :lacks-premium-subscription="!isOptionAvailable('offlineGroup')"
        input-type="percent"
        :default="DefaultValueOfValidatorOptionsNeedingPremium.offlineGroup"
        class="row"
      />
      <NotificationsSubscriptionRow v-model="modifiableOptions.missedAttestations" :t-path="tPath+'missed_attestations'" :lacks-premium-subscription="!isOptionAvailable('missedAttestations')" class="row" />
      <NotificationsSubscriptionRow v-model="modifiableOptions.proposedBlock" :t-path="tPath+'proposed_block'" :lacks-premium-subscription="!isOptionAvailable('proposedBlock')" class="row" />
      <NotificationsSubscriptionRow v-model="modifiableOptions.upcomingProposal" :t-path="tPath+'upcoming_proposal'" :lacks-premium-subscription="!isOptionAvailable('upcomingProposal')" class="row" />
      <NotificationsSubscriptionRow v-model="modifiableOptions.syncCommittee" :t-path="tPath+'sync_committee'" :lacks-premium-subscription="!isOptionAvailable('syncCommittee')" class="row" />
      <NotificationsSubscriptionRow v-model="modifiableOptions.withdrawn" :t-path="tPath+'withdrawn'" :lacks-premium-subscription="!isOptionAvailable('withdrawn')" class="row" />
      <NotificationsSubscriptionRow v-model="modifiableOptions.slashed" :t-path="tPath+'slashed'" :lacks-premium-subscription="!isOptionAvailable('slashed')" class="row" />
      <NotificationsSubscriptionRow v-model="modifiableOptions.realTime" :t-path="tPath+'real_time'" :lacks-premium-subscription="!isOptionAvailable('realTime')" class="row" />
      <hr>
      <NotificationsSubscriptionRow v-model="allCheckbox" :t-path="tPath+'all'" :lacks-premium-subscription="false" class="row" />
    </div>

    <div v-else-if="props.accountSub">
      <NotificationsSubscriptionRow v-model="modifiableOptions.incoming" :t-path="tPath+'incoming'" :lacks-premium-subscription="!isOptionAvailable('incoming')" class="row" />
      <NotificationsSubscriptionRow v-model="modifiableOptions.outgoing" :t-path="tPath+'outgoing'" :lacks-premium-subscription="!isOptionAvailable('outgoing')" class="row" />
      <NotificationsSubscriptionRow v-model="modifiableOptions.erc20" :t-path="tPath+'erc20'" :lacks-premium-subscription="!isOptionAvailable('erc20')" input-type="amount" class="row" />
      <NotificationsSubscriptionRow v-model="modifiableOptions.erc721" :t-path="tPath+'erc721'" :lacks-premium-subscription="!isOptionAvailable('erc721')" class="row" />
      <NotificationsSubscriptionRow v-model="modifiableOptions.erc1155" :t-path="tPath+'erc1155'" :lacks-premium-subscription="!isOptionAvailable('erc1155')" class="row" />
      <hr>
      <NotificationsSubscriptionRow v-model="allCheckbox" :t-path="tPath+'all'" :lacks-premium-subscription="false" class="row" />
      <NotificationsSubscriptionRow v-model="modifiableOptions.networks" :t-path="tPath+'networks'" :lacks-premium-subscription="!isOptionAvailable('networks')" input-type="networks" class="row" />
      <NotificationsSubscriptionRow v-model="modifiableOptions.ignoreSpam" :t-path="tPath+'ignore_spam'" :lacks-premium-subscription="!isOptionAvailable('ignoreSpam')" class="row" />
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
