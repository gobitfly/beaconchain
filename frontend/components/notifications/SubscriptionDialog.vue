<script setup lang="ts">
import { warn } from 'vue'
import type { NotificationSettingsValidatorDashboard, NotificationSettingsAccountDashboard, InternalEntry, APIentry } from '~/types/notifications/subscriptionModal'
import { ChainFamily } from '~/types/network'
import type { ApiErrorResponse } from '~/types/api/common'
import { API_PATH } from '~/types/customFetch'

interface Props {
  dashboardId: number,
  validatorSub?: NotificationSettingsValidatorDashboard,
  accountSub?: NotificationSettingsAccountDashboard
}

type DefinedAPIentry = Exclude<APIentry, null|undefined>
type AllOptions = NotificationSettingsValidatorDashboard & NotificationSettingsAccountDashboard

// #### DIALOG SETTINGS ####

const TimeoutForSavingFailures = 2300 // ms. We cannot let the user close the dialog and later interrupt his/her new activities with "we lost your preferences half a minute ago, we hope you remember them and do not mind going back to that dialog"
const MinimumTimeBetweenAPIcalls = 700 // ms. Any change ends-up saved anyway, so we can prevent useless requests with a delay larger than usual.
const DefaultValues = new Map<keyof AllOptions, DefinedAPIentry>([
  ['group_offline_threshold', -10], // means "10% and unchecked"
  ['is_real_time_mode_enabled', false],
  ['is_erc20_token_transfers_subscribed', false],
  ['erc20_token_transfers_threshold', NaN], // means "empty"
  ['subscribed_chain_ids', []]
])
const orderOfTheRowsInValidatorModal: Array<keyof NotificationSettingsValidatorDashboard | 'ALL'> = [
  'is_validator_offline_subscribed', 'group_offline_threshold', 'is_attestations_missed_subscribed', 'is_block_proposal_subscribed', 'is_upcoming_block_proposal_subscribed',
  'is_sync_subscribed', 'is_withdrawal_processed_subscribed', 'is_slashed_subscribed', 'is_real_time_mode_enabled', 'ALL'
]
const orderOfTheRowsInAccountModal: Array<keyof NotificationSettingsAccountDashboard | 'ALL'> = [
  'is_incoming_transactions_subscribed', 'is_outgoing_transactions_subscribed', 'erc20_token_transfers_threshold', 'is_erc721_token_transfers_subscribed',
  'is_erc1155_token_transfers_subscribed', 'ALL', 'subscribed_chain_ids', 'is_ignore_spam_transactions_enabled'
]
const OptionsOutsideTheScopeOfCheckboxall: Array<keyof(AllOptions)> =
  ['subscribed_chain_ids', 'is_ignore_spam_transactions_enabled'] // options that are not in the group of the all-checkbox
const OptionsNeedingPremium: Array<keyof AllOptions> =
  ['group_offline_threshold', 'is_real_time_mode_enabled']
const RowsThatExpectAPercentage: Array<keyof AllOptions> =
  ['group_offline_threshold']
const RowsWhoseCheckBoxIsInASeparateField = new Map<keyof AllOptions, keyof AllOptions>([
  ['erc20_token_transfers_threshold', 'is_erc20_token_transfers_subscribed']
])

// #### END OF DIALOG SETTINGS ####

type ModifiableOptions = Record<keyof AllOptions, InternalEntry>

const { props, dialogRef } = useBcDialog<Props>({ showHeader: false })
const { t } = useI18n()
const { fetch, setTimeout } = useCustomFetch()
const toast = useBcToast()
const { networkInfo } = useNetworkStore()
const { user } = useUserStore()

const tPath = ref('')
let orderOfTheRows: typeof orderOfTheRowsInValidatorModal | typeof orderOfTheRowsInAccountModal = []
const modifiableOptions = ref({} as ModifiableOptions)
const checkboxAll = ref<InternalEntry>({ type: 'binary', check: false })

const debouncer = useDebounceValue<number>(0, MinimumTimeBetweenAPIcalls)
watch(debouncer.value, sendUserPreferencesToAPI)

let dataNonce = 0 // used by the watcher of `modifiableOptions` to know when it is unnecessary to send changes to the API (it doesn't send if the nonce is 0)

watch(props, (props) => {
  if (!props || (!props.validatorSub && !props.accountSub)) {
    return
  }
  let originalSettings = {} as AllOptions
  if (props.validatorSub) {
    tPath.value = 'notifications.subscriptions.validators.'
    originalSettings = toRaw(props.validatorSub) as AllOptions
    orderOfTheRows = orderOfTheRowsInValidatorModal
  } else {
    tPath.value = 'notifications.subscriptions.accounts.'
    originalSettings = toRaw(props.accountSub!) as AllOptions
    orderOfTheRows = orderOfTheRowsInAccountModal
  }
  modifiableOptions.value = {} as ModifiableOptions
  dataNonce = 0
  for (const key of orderOfTheRows) {
    if (key === 'ALL') { continue }
    modifiableOptions.value[key] = convertAPIentryToInternalEntry(originalSettings, key)
  }
}, { immediate: true })

function convertAPIentryToInternalEntry (apiData: AllOptions, apiKey: keyof AllOptions) : InternalEntry {
  let srcValue = apiData[apiKey]
  if (srcValue === undefined || srcValue === null || (typeof srcValue === 'number' && isNaN(srcValue)) || (Array.isArray(srcValue) && !srcValue.length)) {
    if (DefaultValues.has(apiKey)) {
      srcValue = DefaultValues.get(apiKey)!
    } else {
      warn('Entry `', apiKey, '`is missing in the API data and we do not have a default value for it.')
      return {} as InternalEntry
    }
    dataNonce++ // this will trigger a call to the API to save the settings
  }
  if (Array.isArray(srcValue)) {
    return {
      type: 'networks',
      networks: [...srcValue]
    }
  }
  switch (typeof srcValue) {
    case 'boolean' :
      return {
        type: 'binary',
        check: srcValue
      }
    case 'number' : {
      let check: boolean
      if (RowsWhoseCheckBoxIsInASeparateField.has(apiKey)) {
        check = convertAPIentryToInternalEntry(apiData, RowsWhoseCheckBoxIsInASeparateField.get(apiKey)!).check!
      } else {
        check = srcValue > 0
      }
      return {
        type: (apiKey in RowsThatExpectAPercentage) ? 'percent' : 'amount',
        check,
        num: Math.abs(srcValue)
      }
    }
  }
}

function checkboxAllHasBeenClicked (checked: boolean) : void {
  for (const k of Object.keys(modifiableOptions.value)) {
    const key = k as keyof ModifiableOptions
    if (isOptionAvailable(key) && !OptionsOutsideTheScopeOfCheckboxall.includes(key)) {
      modifiableOptions.value[key].check = checked
    }
  }
  // no need to call the API, the modifications that we did in `modifiableOptions.value` will trigger its watcher (that calls the API)
}

watch(modifiableOptions, (options) => {
  checkboxAll.value.check = true
  for (const k in options) {
    const key = k as keyof ModifiableOptions
    if (isOptionAvailable(key) && !OptionsOutsideTheScopeOfCheckboxall.includes(key)) {
      checkboxAll.value.check &&= options[key].check
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
  const output = {} as Record<string, APIentry>
  for (const k in modifiableOptions.value) {
    const key = k as keyof ModifiableOptions
    const value = toRaw(modifiableOptions.value[key])
    switch (value.type) {
      case 'binary' :
        output[key] = value.check
        break
      case 'amount' :
      case 'percent' :
        output[key] = isNaN(value.num!) ? null : (value.check ? value.num : -value.num!)
        if (RowsWhoseCheckBoxIsInASeparateField.has(key)) {
          output[RowsWhoseCheckBoxIsInASeparateField.get(key)!] = !isNaN(value.num!) && value.check
        }
        break
      case 'networks' :
        output[key] = value.networks
        break
    }
  }
  // now we send the data
  let response: ApiErrorResponse | undefined
  try {
    setTimeout(TimeoutForSavingFailures)
    response = await fetch<ApiErrorResponse>(API_PATH.SETTINGS_DASHBOARDS, {
      method: 'POST',
      body: output
    }, {
      for: props.value!.accountSub ? 'accounts' : 'validators',
      dashboardKey: String(props.value?.dashboardId)
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

const isOptionAvailable = (key: keyof AllOptions) => user.value?.premium_perks.ad_free || !OptionsNeedingPremium.includes(key)
</script>

<template>
  <div v-if="props && tPath" class="content">
    <div class="title">
      {{ t('notifications.subscriptions.dialog_title') }}
    </div>

    <div v-if="t(tPath+'explanation')" class="explanation">
      {{ t(tPath+'explanation', (networkInfo.family === ChainFamily.Gnosis) ? 5 : 20) }}
    </div>

    <div v-for="row of orderOfTheRows" :key="row" class="row-container">
      <NotificationsSubscriptionRow
        v-if="row != 'ALL'"
        v-model="modifiableOptions[row]"
        :t-path="tPath+row"
        :lacks-premium-subscription="!isOptionAvailable(row)"
        :value-in-text="(row == 'is_attestations_missed_subscribed') ? Math.round(networkInfo.secondsPerSlot*networkInfo.slotsPerEpoch/6)/10 : undefined"
        class="row"
      />
      <div v-if="row == 'ALL'" class="separation" />
      <NotificationsSubscriptionRow
        v-if="row == 'ALL'"
        v-model="checkboxAll"
        :t-path="tPath+'all'"
        :lacks-premium-subscription="false"
        class="row"
        @checkbox-click="checkboxAllHasBeenClicked"
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
