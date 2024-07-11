<script setup lang="ts">
import type { NotificationEventsValidatorDashboard, NotificationEventsAccountDashboard, CheckboxAndNumber, InputRow } from '~/types/notifications/subscriptionModal'
import type { ChainIDs } from '~/types/network'
import type { ApiErrorResponse } from '~/types/api/common'
import { API_PATH } from '~/types/customFetch'

interface Props {
  validatorSub?: NotificationEventsValidatorDashboard,
  accountSub?: NotificationEventsAccountDashboard
}

// #### DIALOG SETTINGS ####

const TimeoutForSavingFailures = 2300 // ms. We cannot let the user close the dialog and later interrupt his/her new activities with "we lost your preferences half a minute ago, we hope you remember them and do not mind going back to that dialog"
const MinimumTimeBetweenAPIcalls = 700 // ms. Any change ends-up saved anyway, so we can prevent useless requests with a delay larger than usual.
const DefaultValueOfValidatorOptionsNeedingPremium = {
  group_offline: -10, // means "10% and unchecked"
  realtime_mode: false
  // ... add lines here to make options available to premium accounts only
}
const DefaultValueOfAccountOptionsNeedingPremium = {
  // add lines here to make options available to premium accounts only
}
const orderOfTheRowsInValidatorModal: Array<keyof NotificationEventsValidatorDashboard | 'ALL'> =
  ['validator_offline', 'group_offline', 'attestations_missed', 'block_proposal', 'upcoming_block_proposal', 'sync', 'withdrawal_processed', 'slashed', 'realtime_mode', 'ALL']
const orderOfTheRowsInAccountModal: Array<keyof NotificationEventsAccountDashboard | 'ALL'> =
  ['incoming_transactions', 'outgoing_transactions', 'track_erc20_token_transfers', 'track_erc721_token_transfers', 'track_erc1155_token_transfers', 'ALL', 'networks', 'ignore_spam_transactions']
const OptionsOutsideTheScopeOfCheckboxall: Array<keyof(NotificationEventsValidatorDashboard & NotificationEventsAccountDashboard)> =
  ['networks', 'ignore_spam_transactions'] // options that are not in the group of the all-checkbox
const RowsThatExpectAnAmount: Array<keyof(NotificationEventsValidatorDashboard & NotificationEventsAccountDashboard)> =
  ['track_erc20_token_transfers']
const RowsThatExpectAPercentage: Array<keyof(NotificationEventsValidatorDashboard & NotificationEventsAccountDashboard)> =
  ['group_offline']
const RowsThatExpectANetwork: Array<keyof(NotificationEventsValidatorDashboard & NotificationEventsAccountDashboard)> =
  ['networks']

// #### END OF DIALOG SETTINGS ####

type AllOptions = NotificationEventsValidatorDashboard & NotificationEventsAccountDashboard & typeof DefaultValueOfValidatorOptionsNeedingPremium & typeof DefaultValueOfAccountOptionsNeedingPremium
type ModifiableOptions = Record<keyof AllOptions, CheckboxAndNumber|ChainIDs[]>

const { props, dialogRef } = useBcDialog<Props>({ showHeader: false })
const { t } = useI18n()
const { fetch, setTimeout } = useCustomFetch()
const toast = useBcToast()
const { networkInfo } = useNetworkStore()
const { user } = useUserStore()

const tPath = ref('')
let orderOfTheRows: typeof orderOfTheRowsInValidatorModal | typeof orderOfTheRowsInAccountModal = []
let originalSettings = {} as AllOptions
const modifiableOptions = ref({} as ModifiableOptions)
const checkboxAll = ref({ check: false } as CheckboxAndNumber)

const debouncer = useDebounceValue<number>(0, MinimumTimeBetweenAPIcalls)
watch(debouncer.value, sendUserPreferencesToAPI)

let dataNonce = 0 // used by the watcher of `modifiableOptions` to know when it is unnecessary to send apparent changes to the API (it doesn't send if the nonce is 0)

watch(props, (props) => {
  if (!props || (!props.validatorSub && !props.accountSub)) {
    return
  }
  if (props.validatorSub) {
    tPath.value = 'notifications.subscriptions.validators.'
    originalSettings = { ...DefaultValueOfValidatorOptionsNeedingPremium, ...structuredClone(toRaw(props.validatorSub)) } as AllOptions
    orderOfTheRows = orderOfTheRowsInValidatorModal
  } else {
    tPath.value = 'notifications.subscriptions.accounts.'
    originalSettings = { ...DefaultValueOfAccountOptionsNeedingPremium, ...structuredClone(toRaw(props.accountSub)) } as AllOptions
    orderOfTheRows = orderOfTheRowsInAccountModal
  }
  modifiableOptions.value = {} as ModifiableOptions
  for (const entry of Object.entries(originalSettings)) {
    const key = entry[0] as keyof typeof originalSettings
    if (Array.isArray(entry[1])) {
      modifiableOptions.value[key] = entry[1]
    } else {
      switch (typeof entry[1]) {
        case 'boolean' :
          modifiableOptions.value[key] = {
            check: entry[1],
            num: 0
          }
          break
        default :
          modifiableOptions.value[key] = {
            check: (entry[1] != null && entry[1] >= 0),
            num: (entry[1] === null) ? null : Math.abs(entry[1])
          }
          break
      }
    }
  }
  dataNonce = 0
}, { immediate: true })

function checkboxAllhasBeenClicked (checked: boolean) : void {
  for (const k of Object.keys(modifiableOptions.value)) {
    const key = k as keyof ModifiableOptions
    if (isOptionAvailable(key) && !OptionsOutsideTheScopeOfCheckboxall.includes(key)) {
      (modifiableOptions.value[key] as CheckboxAndNumber).check = checked
    }
  }
  // no need to call the API, the modifications that we did in `modifiableOptions.value` will trigger its watcher (that calls the API)
}

watch(modifiableOptions, (options) => {
  checkboxAll.value.check = true
  for (const k of Object.keys(options)) {
    const key = k as keyof ModifiableOptions
    if (isOptionAvailable(key) && !OptionsOutsideTheScopeOfCheckboxall.includes(key)) {
      checkboxAll.value.check &&= (options[key] as CheckboxAndNumber).check
    }
  }
  if (dataNonce > 0) {
    // it will call `sendUserPreferencesToAPI()`
    debouncer.bounce(dataNonce, false, true)
  }
  dataNonce++
}, { immediate: true, deep: true })

async function sendUserPreferencesToAPI () {
  // first we convert our internal structures to the format of the API
  const output = {} as Record<string, any>
  for (const entry of Object.entries(modifiableOptions.value)) {
    const key = entry[0] as keyof ModifiableOptions
    if (Array.isArray(entry[1])) {
      output[key] = entry[1]
    } else {
      switch (typeof originalSettings[key]) {
        case 'boolean' :
          output[key] = entry[1].check
          break
        default :
          output[key] = entry[1].num
          if (!entry[1].check && entry[1].num !== null) { output[key] *= -1 }
          break
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
        category: props.value!.accountSub ? 'accounts' : 'validators',
        subscriptions: output
      }
    })
  } catch {
    response = undefined
  }
  if (!response || response.error) {
    toast.showError({ summary: t('notifications.subscriptions.error_title'), group: t('notifications.subscriptions.error_group'), detail: t('notifications.subscriptions.error_message') })
  }
}

function closeDialog () : void {
  dialogRef?.value.close()
}

const isOptionAvailable = (key: keyof AllOptions) => !user.value?.premium_perks.ad_free || !(key in DefaultValueOfValidatorOptionsNeedingPremium || key in DefaultValueOfAccountOptionsNeedingPremium)

function getRowType (key: keyof AllOptions) : InputRow {
  if (RowsThatExpectAnAmount.includes(key)) { return 'amount' }
  if (RowsThatExpectAPercentage.includes(key)) { return 'percent' }
  if (RowsThatExpectANetwork.includes(key)) { return 'networks' }
  return 'binary'
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

    <div v-for="row of orderOfTheRows" :key="row" class="row-container">
      <NotificationsSubscriptionRow
        v-if="row != 'ALL'"
        v-model="modifiableOptions[row]"
        :t-path="tPath+row"
        :lacks-premium-subscription="!isOptionAvailable(row)"
        :input-type="getRowType(row)"
        :value-in-text="(row == 'attestations_missed') ? Math.round(networkInfo.secondsPerSlot*networkInfo.slotsPerEpoch/6)/10 : undefined"
        class="row"
      />
      <div v-if="row == 'ALL'" class="separation" />
      <NotificationsSubscriptionRow
        v-if="row == 'ALL'"
        v-model="checkboxAll"
        :t-path="tPath+'all'"
        :lacks-premium-subscription="false"
        input-type="binary"
        class="row"
        @checkbox-click="checkboxAllhasBeenClicked"
      />
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
    color: var(--text-color-discreet)
  }

  .row-container {
    position: relative;
    margin-top: 8px;
    margin-bottom: 8px;
    .separation {
      height: 1px;
      background-color: var(--container-border-color);
      margin-bottom: 16px;
    }
  }

  .footer {
    display: flex;
    justify-content: right;
    margin-top: var(--padding);
    gap: var(--padding);
  }
}
</style>
