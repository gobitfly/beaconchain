<script setup lang="ts">
import { BcDialogConfirm, NotificationsSubscriptionDialog } from '#components'
import type { NotificationSettingsValidatorDashboard, NotificationSettingsAccountDashboard } from '~/types/api/notifications'

const dialog = useDialog()

function onClose (answer: boolean) {
  setTimeout(() => {
    alert('response: ' + answer)
  }, 100
  )
}

const openQuestion = (yesLabel?: string, noLabel?: string) => {
  dialog.open(BcDialogConfirm, {
    data: {
      title: 'My super question',
      question: 'Are you ready to rumble, or do you have second thoughts?',
      yesLabel,
      noLabel
    },
    onClose: response => onClose(response?.data)
  })
}

const validatorSub: NotificationSettingsValidatorDashboard = {
  webhook_url: 'http://bablabla',
  is_webhook_discord_enabled: true,
  is_validator_offline_subscribed: true,
  group_offline_threshold: 0, // means "deactivated/unchecked"
  is_attestations_missed_subscribed: true,
  is_block_proposal_subscribed: true,
  is_upcoming_block_proposal_subscribed: false,
  is_sync_subscribed: true,
  is_withdrawal_processed_subscribed: true,
  is_slashed_subscribed: false,
  is_real_time_mode_enabled: false
}

const accountSub: NotificationSettingsAccountDashboard = {
  webhook_url: 'http://bablabla',
  is_webhook_discord_enabled: true,
  is_incoming_transactions_subscribed: true,
  is_outgoing_transactions_subscribed: true,
  is_erc20_token_transfers_subscribed: false,
  erc20_token_transfers_value_threshold: 0,
  is_erc721_token_transfers_subscribed: true,
  is_erc1155_token_transfers_subscribed: false,
  subscribed_chain_ids: [17000],
  is_ignore_spam_transactions_enabled: true
}

function openSubscriptions (props: any) {
  dialog.open(NotificationsSubscriptionDialog, {
    data: props,
    onClose: response => onClose(response?.data)
  })
}
</script>

<template>
  <div class="container">
    Note: to test the saving of the options to the API, open the dialogs from the notification dashboard.<br>The communication with the API is implemented there.
    <Button @click="openSubscriptions({ dashboardType: 'validator', initialSettings: validatorSub, saveUserSettings: () => {} })">
      Subscribe to notifications for your validators
    </Button>
    <Button @click="openSubscriptions({ dashboardType: 'account', initialSettings: accountSub, saveUserSettings: () => {} })">
      Subscribe to notifications for your accounts
    </Button>
    <br>
    <Button @click="openQuestion()">
      Open Question
    </Button>
    <Button @click="openQuestion('Are you sure')">
      Open Question sure?
    </Button>
    <Button @click="openQuestion(undefined, 'cancel')">
      Open Question cancel?
    </Button>
  </div>
</template>

<style lang="scss" scoped>
.container{
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 10px;
  padding: 10px;
}
</style>
