<script setup lang="ts">
import { warn } from 'vue'
import type {
  InternalEntry,
  APIentry,
} from '~/types/notifications/subscriptionModal'
import type {
  NotificationSettingsValidatorDashboard,
  NotificationSettingsAccountDashboard,
} from '~/types/api/notifications'
import { ChainFamily } from '~/types/network'
import type { DashboardType } from '~/types/dashboard'

type AllOptions = NotificationSettingsValidatorDashboard &
  NotificationSettingsAccountDashboard
type DefinedAPIentry = Exclude<APIentry, null | undefined>

interface Props {
  dashboardType: DashboardType
  initialSettings: AllOptions
  saveUserSettings: (
    settings: Record<keyof AllOptions, DefinedAPIentry>,
  ) => void
}

// #### CONFIGURATION OF THE DIALOGS ####

const DefaultValues = new Map<keyof AllOptions, InternalEntry>([
  ['group_offline_threshold', { type: 'percent', check: false, num: 10 }],
  ['is_real_time_mode_enabled', { type: 'binary', check: false }],
  [
    'erc20_token_transfers_value_threshold',
    { type: 'amount', check: false, num: NaN },
  ], // NaN will leave the input field empty (the user sees the placeholder)
  ['subscribed_chain_ids', { type: 'networks', networks: [] }],
])
const orderOfTheRowsInValidatorModal: Array<
  keyof NotificationSettingsValidatorDashboard | 'ALL'
> = [
  'is_validator_offline_subscribed',
  'group_offline_threshold',
  'is_attestations_missed_subscribed',
  'is_block_proposal_subscribed',
  'is_upcoming_block_proposal_subscribed',
  'is_sync_subscribed',
  'is_withdrawal_processed_subscribed',
  'is_slashed_subscribed',
  'is_real_time_mode_enabled',
  'ALL',
]
const orderOfTheRowsInAccountModal: Array<
  keyof NotificationSettingsAccountDashboard | 'ALL'
> = [
  'is_incoming_transactions_subscribed',
  'is_outgoing_transactions_subscribed',
  'erc20_token_transfers_value_threshold',
  'is_erc721_token_transfers_subscribed',
  'is_erc1155_token_transfers_subscribed',
  'ALL',
  'subscribed_chain_ids',
  'is_ignore_spam_transactions_enabled',
]
const RowsWhoseCheckBoxIsInASeparateField = new Map<
  keyof AllOptions,
  keyof AllOptions
>([
  [
    'erc20_token_transfers_value_threshold',
    'is_erc20_token_transfers_subscribed',
  ],
])
const OptionsOutsideTheScopeOfCheckboxall: Array<keyof AllOptions> = [
  'subscribed_chain_ids',
  'is_ignore_spam_transactions_enabled',
] // options that are not in the group of the all-checkbox
const OptionsNeedingPremium: Array<keyof AllOptions> = [
  'group_offline_threshold',
  'is_real_time_mode_enabled',
]
const RowsThatExpectAPercentage: Array<keyof AllOptions> = [
  'group_offline_threshold',
]

// #### END OF CONFIGURATION OF THE DIALOGS ####

type ModifiableOptions = Record<keyof AllOptions, InternalEntry>

const { props, dialogRef } = useBcDialog<Props>({ showHeader: false })
const { t } = useTranslation()
const { networkInfo } = useNetworkStore()
const { user } = useUserStore()

const tPath = ref('')
let orderOfTheRows:
  | typeof orderOfTheRowsInValidatorModal
  | typeof orderOfTheRowsInAccountModal = []
let originalSettings: AllOptions
const modifiableOptions = ref({} as ModifiableOptions)
const checkboxAll = ref<InternalEntry>({ type: 'binary', check: false })

// used by the watcher of `modifiableOptions` to know when it is unnecessary
// to send changes to the API (it doesn't send if the nonce is 0)
let dataNonce = 0

const getOptionType = (key: keyof AllOptions) =>
  Array.isArray(originalSettings[key])
    ? 'networks'
    : typeof originalSettings[key] === 'boolean'
      ? 'binary'
      : RowsThatExpectAPercentage.includes(key)
        ? 'percent'
        : 'amount'
const isOptionValueKnownInDB = (key: keyof AllOptions) =>
  originalSettings[key] !== undefined
  && originalSettings[key] !== null
  && (typeof originalSettings[key] !== 'number'
  || (originalSettings[key] as number) > 0
  || isOptionActivatedInDB(key))
  && (!Array.isArray(originalSettings[key])
  || !!(originalSettings[key] as Array<any>).length)
const isOptionActivatedInDB = (key: keyof AllOptions) =>
  RowsWhoseCheckBoxIsInASeparateField.has(key)
    ? !!originalSettings[RowsWhoseCheckBoxIsInASeparateField.get(key)!]
    : !!originalSettings[key]
const isOptionAvailable = (key: keyof AllOptions) =>
  user.value?.premium_perks.ad_free || !OptionsNeedingPremium.includes(key)

watch(
  props,
  (props) => {
    if (!props || !props.initialSettings) {
      return
    }
    originalSettings = toRaw(props.initialSettings)
    switch (props.dashboardType) {
      case 'validator':
        tPath.value = 'notifications.subscriptions.validators.'
        orderOfTheRows = orderOfTheRowsInValidatorModal
        break
      case 'account':
        tPath.value = 'notifications.subscriptions.accounts.'
        orderOfTheRows = orderOfTheRowsInAccountModal
        break
      default:
        return
    }
    modifiableOptions.value = {} as ModifiableOptions
    dataNonce = 0
    for (const key of orderOfTheRows) {
      if (key === 'ALL') continue
      modifiableOptions.value[key] = convertAPIentryToInternalEntry(key)
    }
  },
  { immediate: true },
)

function checkboxAllHasBeenClicked(checked: boolean): void {
  for (const k of Object.keys(modifiableOptions.value)) {
    const key = k as keyof ModifiableOptions
    if (
      isOptionAvailable(key)
      && !OptionsOutsideTheScopeOfCheckboxall.includes(key)
    ) {
      modifiableOptions.value[key].check = checked
    }
  }
}

watch(
  modifiableOptions,
  (options) => {
    checkboxAll.value.check = true
    for (const k in options) {
      const key = k as keyof ModifiableOptions
      if (
        isOptionAvailable(key)
        && !OptionsOutsideTheScopeOfCheckboxall.includes(key)
      ) {
        checkboxAll.value.check &&= options[key].check
      }
    }
    if (dataNonce > 0) {
      sendUserPreferencesToAPI()
    }
    dataNonce++
  },
  { immediate: true, deep: true },
)

/** reads data that our parent received from the API and converts it to our internal format */
function convertAPIentryToInternalEntry(
  apiKey: keyof AllOptions,
): InternalEntry {
  const srcValue = originalSettings[apiKey]
  const type = getOptionType(apiKey)
  if (!isOptionValueKnownInDB(apiKey)) {
    if (DefaultValues.has(apiKey)) {
      return { ...DefaultValues.get(apiKey)! }
    }
    else {
      warn(
        'A value for entry `'
        + apiKey
        + '` is not in the the database and the front-end does not have a default value for it.',
      )
      return {} as InternalEntry
    }
  }
  switch (type) {
    case 'networks':
      return {
        type,
        networks: [...(srcValue as number[])],
      }
    case 'binary':
      return {
        type,
        check: srcValue as boolean,
      }
    case 'percent':
      return {
        type,
        check: isOptionActivatedInDB(apiKey),
        num: (srcValue as number) * 100,
      }
    case 'amount':
      return {
        type,
        check: isOptionActivatedInDB(apiKey),
        num: srcValue as number,
      }
  }
}

/** converts our internal data to the format understood by the API and sends it */
function sendUserPreferencesToAPI() {
  // conversion
  const output = {} as Record<keyof AllOptions, DefinedAPIentry>
  for (const k in modifiableOptions.value) {
    const key = k as keyof ModifiableOptions
    const value = toRaw(modifiableOptions.value[key])
    switch (value.type) {
      case 'binary':
        output[key] = value.check!
        break
      case 'percent':
      case 'amount': {
        const num = value.type === 'percent' ? value.num! / 100 : value.num!
        const activate = !isNaN(num) && value.check!
        if (RowsWhoseCheckBoxIsInASeparateField.has(key)) {
          output[key] = !isNaN(num) ? num : 0
          output[RowsWhoseCheckBoxIsInASeparateField.get(key)!] = activate
        }
        else {
          output[key] = activate ? num : 0
        }
        break
      }
      case 'networks':
        output[key] = value.networks!
        break
    }
  }
  // sending
  props.value?.saveUserSettings(output)
}

function closeDialog(): void {
  dialogRef?.value.close()
}
</script>

<template>
  <div
    v-if="props && tPath"
    class="content"
  >
    <div class="title">
      {{ t("notifications.subscriptions.dialog_title") }}
    </div>

    <div
      v-if="t(tPath + 'explanation')"
      class="explanation"
    >
      {{
        t(
          tPath + "explanation",
          networkInfo.family === ChainFamily.Gnosis ? 5 : 20,
        )
      }}
    </div>

    <div
      v-for="row of orderOfTheRows"
      :key="row"
      class="row-container"
    >
      <NotificationsManagementSubscriptionRow
        v-if="row != 'ALL'"
        v-model="modifiableOptions[row]"
        :t-path="tPath + row"
        :lacks-premium-subscription="!isOptionAvailable(row)"
        :value-in-text="
          row == 'is_attestations_missed_subscribed'
            ? Math.round(
              (networkInfo.secondsPerSlot * networkInfo.slotsPerEpoch) / 6,
            ) / 10
            : undefined
        "
        class="row"
      />
      <div
        v-if="row == 'ALL'"
        class="separation"
      />
      <NotificationsManagementSubscriptionRow
        v-if="row == 'ALL'"
        v-model="checkboxAll"
        :t-path="tPath + 'all'"
        :lacks-premium-subscription="false"
        class="row"
        @checkbox-click="checkboxAllHasBeenClicked"
      />
    </div>

    <div class="footer">
      <Button
        type="button"
        :label="t('notifications.subscriptions.button')"
        @click="closeDialog"
      />
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/fonts.scss";

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
    color: var(--text-color-discreet);
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
