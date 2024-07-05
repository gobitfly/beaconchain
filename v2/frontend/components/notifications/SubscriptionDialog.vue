<script setup lang="ts">
import type { ValidatorSubscriptionState, AccountSubscriptionState, CheckboxAndNumber } from '~/types/subscriptionModal'
import type { ChainID } from '~/types/network'
import type { ApiErrorResponse } from '~/types/api/common'
import { API_PATH } from '~/types/customFetch'

interface Props {
  validatorSub?: ValidatorSubscriptionState,
  accountSub?: AccountSubscriptionState,
  premiumUser: boolean
}

const TimeoutForSavingFailures = 2000 // ms. We cannot let the user close the dialog and later interrupt his/her new activities with "we lost what you did half a minute ago, we hope you remember your preferences and do not mind going back to that dialog"
const MinimumTimeBetweenAPIcalls = 700 // ms
const DefaultValueOfValidatorOptionsNeedingPremium = {
  offlineGroup: -10, // means "10% and unchecked"
  realTime: false
  // ... add lines here to make options available to premium accounts only
}
const DefaultValueOfAccountOptionsNeedingPremium = {
  // add lines here to make options available to premium accounts only
}
const OptionsOutsideTheScopeOfCheckboxall: Array<keyof(ValidatorSubscriptionState & AccountSubscriptionState)> =
      ['networks', 'ignoreSpam'] // options that are not in the group of the all-checkbox

type AllPossibleOptions = ValidatorSubscriptionState & AccountSubscriptionState & typeof DefaultValueOfValidatorOptionsNeedingPremium & typeof DefaultValueOfAccountOptionsNeedingPremium
type ModifiableOptions = Record<keyof AllPossibleOptions, CheckboxAndNumber|ChainID[]>

const { props, dialogRef } = useBcDialog<Props>({ showHeader: false })
const { t } = useI18n()
const { fetch, setTimeout } = useCustomFetch()
const toast = useBcToast()

const tPath = ref('')
let originalSettings = {} as AllPossibleOptions
const modifiableOptions = ref({} as ModifiableOptions)
const allCheckbox = ref({ check: false } as CheckboxAndNumber)

const debouncer = useDebounceValue<number>(0, MinimumTimeBetweenAPIcalls)
watch(debouncer.value, sendUserPreferencesToAPI)

let dataNonce = 0 // is used by the watcher of `modifiableOptions` to know when it is unnecessary to send apparent changes to the API (it doesn't send if the nonce is 0)

watch(props, (props) => {
  if (!props || (!props.validatorSub && !props.accountSub)) {
    return
  }
  if (props.validatorSub) {
    tPath.value = 'notifications.subscriptions.validators.'
    originalSettings = { ...DefaultValueOfValidatorOptionsNeedingPremium, ...structuredClone(toRaw(props.validatorSub)) } as AllPossibleOptions
  } else {
    tPath.value = 'notifications.subscriptions.accounts.'
    originalSettings = { ...DefaultValueOfAccountOptionsNeedingPremium, ...structuredClone(toRaw(props.accountSub)) } as AllPossibleOptions
  }
  modifiableOptions.value = {} as ModifiableOptions
  for (const entry of Object.entries(originalSettings)) {
    const key = entry[0] as keyof typeof originalSettings
    if (Array.isArray(entry[1])) {
      modifiableOptions.value[key] = entry[1]
    } else {
      switch (typeof entry[1]) {
        case 'boolean' : modifiableOptions.value[key] = { check: entry[1], num: 0 }; break
        case 'number' : modifiableOptions.value[key] = { check: entry[1] >= 0, num: Math.abs(entry[1]) }; break
      }
    }
  }
  dataNonce = 0
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
  for (const k of Object.keys(options)) {
    const key = k as keyof ModifiableOptions
    if (isOptionAvailable(key) && !OptionsOutsideTheScopeOfCheckboxall.includes(key)) {
      allCheckbox.value.check &&= (options[key] as CheckboxAndNumber).check
    }
  }
  if (dataNonce > 0) {
    // it will call `sendUserPreferencesToAPI()`
    debouncer.bounce(dataNonce, false, true)
  }
  dataNonce++
}, { immediate: true, deep: true })

let lastSaveFailed = false

async function sendUserPreferencesToAPI () {
  // first we convert our internal structures to the format of the API
  const output = {} as Record<string, any>
  for (const entry of Object.entries(modifiableOptions.value)) {
    const key = entry[0] as keyof ModifiableOptions
    const snaKey = camelToSnakeCase(key)
    if (Array.isArray(entry[1])) {
      output[snaKey] = entry[1]
    } else {
      switch (typeof originalSettings[key]) {
        case 'boolean' : output[snaKey] = entry[1].check; break
        case 'number' : output[snaKey] = entry[1].num * (entry[1].check ? 1 : -1); break
      }
    }
  }
  // now we send the data
  let response: ApiErrorResponse | undefined
  try {
    setTimeout(TimeoutForSavingFailures)
    response = await fetch<ApiErrorResponse>(API_PATH.NOTIFICATION_SUBSCRIPTIONS, {
      method: 'POST',
      body: {
        category: props.value?.accountSub ? 'accounts' : 'validators',
        subscriptions: output
      }
    })
  } catch {
    response = undefined
  }
  if (!response || response.error) {
    toast.showError({ summary: t('notifications.subscriptions.error_title'), group: t('notifications.subscriptions.error_group'), detail: t('notifications.subscriptions.error_message') })
    lastSaveFailed = true // we will try again when the dialog closes
  } else {
    lastSaveFailed = false
  }
}

const isOptionAvailable = (key: string) => props.value?.premiumUser || !(key in DefaultValueOfValidatorOptionsNeedingPremium || key in DefaultValueOfAccountOptionsNeedingPremium)

const closeDialog = () => {
  if (lastSaveFailed) {
    // second chance: we try not lose what the user has set
    debouncer.bounce(dataNonce, false, true)
  }
  dialogRef?.value.close(true)
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
      <Button type="button" :label="t('notifications.subscriptions.button')" @click="closeDialog" />
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
