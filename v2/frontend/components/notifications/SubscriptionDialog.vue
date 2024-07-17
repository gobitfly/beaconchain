<script setup lang="ts">
import { warn } from 'vue'
import type { NotificationSettingsValidatorDashboard, NotificationSettingsAccountDashboard, InternalEntry, APIentry } from '~/types/notifications/subscriptionModal'
import { ChainFamily } from '~/types/network'
import type { ApiErrorResponse } from '~/types/api/common'
import { API_PATH } from '~/types/customFetch'

interface Props {
  dashboardId: number,
  groupId: number,
  validatorSub?: NotificationSettingsValidatorDashboard,
  accountSub?: NotificationSettingsAccountDashboard
}

type DefinedAPIentry = Exclude<APIentry, null|undefined>
type AllOptions = NotificationSettingsValidatorDashboard & NotificationSettingsAccountDashboard

// #### DIALOG SETTINGS ####

const TimeoutForSavingFailures = 2300 // ms. We cannot let the user close the dialog and later interrupt his/her new activities with "we lost your preferences half a minute ago, we hope you remember them and do not mind going back to that dialog"
const MinimumTimeBetweenAPIcalls = 700 // ms. Any change ends-up saved anyway, so we can prevent useless requests with a delay larger than usual.
const DefaultValues = new Map<keyof AllOptions, InternalEntry>([
  ['group_offline_threshold', { type: 'percent', check: false, num: 10 }],
  ['is_real_time_mode_enabled', { type: 'binary', check: false }],
  ['erc20_token_transfers_threshold', { type: 'amount', check: false, num: NaN }], // NaN will leave the input field empty (the user sees the placeholder)
  ['subscribed_chain_ids', { type: 'networks', networks: [] }]
])
const orderOfTheRowsInValidatorModal: Array<keyof NotificationSettingsValidatorDashboard | 'ALL'> = [
  'is_validator_offline_subscribed', 'group_offline_threshold', 'is_attestations_missed_subscribed', 'is_block_proposal_subscribed', 'is_upcoming_block_proposal_subscribed',
  'is_sync_subscribed', 'is_withdrawal_processed_subscribed', 'is_slashed_subscribed', 'is_real_time_mode_enabled', 'ALL'
]
const orderOfTheRowsInAccountModal: Array<keyof NotificationSettingsAccountDashboard | 'ALL'> = [
  'is_incoming_transactions_subscribed', 'is_outgoing_transactions_subscribed', 'erc20_token_transfers_threshold', 'is_erc721_token_transfers_subscribed',
  'is_erc1155_token_transfers_subscribed', 'ALL', 'subscribed_chain_ids', 'is_ignore_spam_transactions_enabled'
]
const RowsWhoseCheckBoxIsInASeparateField = new Map<keyof AllOptions, keyof AllOptions>([
  ['erc20_token_transfers_threshold', 'is_erc20_token_transfers_subscribed']
])
const OptionsOutsideTheScopeOfCheckboxall: Array<keyof(AllOptions)> =
  ['subscribed_chain_ids', 'is_ignore_spam_transactions_enabled'] // options that are not in the group of the all-checkbox
const OptionsNeedingPremium: Array<keyof AllOptions> =
  ['group_offline_threshold', 'is_real_time_mode_enabled']
const RowsThatExpectAPercentage: Array<keyof AllOptions> =
  ['group_offline_threshold']

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

function checkboxAllHasBeenClicked (checked: boolean) : void {
  for (const k of Object.keys(modifiableOptions.value)) {
    const key = k as keyof ModifiableOptions
    if (isOptionAvailable(key) && !OptionsOutsideTheScopeOfCheckboxall.includes(key)) {
      modifiableOptions.value[key].check = checked
    }
  }
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

/** reads data that our parent received from the API and converts it to our internal format */
function convertAPIentryToInternalEntry (apiData: AllOptions, apiKey: keyof AllOptions) : InternalEntry {
  const srcValue = apiData[apiKey]
  const type = getOptionType(apiData, apiKey)
  if (!isOptionValueKnownInDB(apiData, apiKey)) {
    if (DefaultValues.has(apiKey)) {
      return { ...DefaultValues.get(apiKey)! }
    } else {
      warn('A value for entry `' + apiKey + '` is not in the the database and the front-end does not have a default value for it.')
      return {} as InternalEntry
    }
  }
  switch (type) {
    case 'networks' :
      return {
        type,
        networks: [...srcValue as number[]]
      }
    case 'binary' :
      return {
        type,
        check: srcValue as boolean
      }
    case 'percent' :
      return {
        type,
        check: isOptionActivatedInDB(apiData, apiKey),
        num: srcValue as number * 100
      }
    case 'amount' :
      return {
        type,
        check: isOptionActivatedInDB(apiData, apiKey),
        num: srcValue as number
      }
  }
}

/** converts our internal data to the format understood by the API and sends it */
async function sendUserPreferencesToAPI () {
  // conversion
  const output = {} as Record<string, DefinedAPIentry>
  for (const k in modifiableOptions.value) {
    const key = k as keyof ModifiableOptions
    const value = toRaw(modifiableOptions.value[key])
    switch (value.type) {
      case 'binary' :
        output[key] = value.check!
        break
      case 'percent' :
      case 'amount' : {
        const num = (value.type === 'percent') ? value.num! / 100 : value.num!
        const activate = !isNaN(num) && value.check!
        if (RowsWhoseCheckBoxIsInASeparateField.has(key)) {
          output[key] = !isNaN(num) ? num : 0
          output[RowsWhoseCheckBoxIsInASeparateField.get(key)!] = activate
        } else {
          output[key] = activate ? num : 0
        }
        break
      }
      case 'networks' :
        output[key] = value.networks!
        break
    }
  }
  // sending
  let response: ApiErrorResponse | undefined
  try {
    setTimeout(TimeoutForSavingFailures)
    response = await fetch<ApiErrorResponse>(API_PATH.SETTINGS_DASHBOARDS, {
      method: 'POST',
      body: output
    }, {
      for: props.value!.accountSub ? 'accounts' : 'validators',
      dashboardKey: String(props.value?.dashboardId),
      groupId: String(props.value?.groupId)
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

const isOptionValueKnownInDB = (apiData: AllOptions, key: keyof AllOptions) =>
  apiData[key] !== undefined && apiData[key] !== null &&
  (typeof apiData[key] !== 'number' || apiData[key] as number > 0 || isOptionActivatedInDB(apiData, key)) &&
  (!Array.isArray(apiData[key]) || !!(apiData[key] as Array<any>).length)
const isOptionActivatedInDB = (apiData: AllOptions, key: keyof AllOptions) => RowsWhoseCheckBoxIsInASeparateField.has(key) ? !!apiData[RowsWhoseCheckBoxIsInASeparateField.get(key)!] : !!apiData[key]
const getOptionType = (apiData: AllOptions, key: keyof AllOptions) => Array.isArray(apiData[key]) ? 'networks' : (typeof apiData[key] === 'boolean' ? 'binary' : (RowsThatExpectAPercentage.includes(key) ? 'percent' : 'amount'))
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
