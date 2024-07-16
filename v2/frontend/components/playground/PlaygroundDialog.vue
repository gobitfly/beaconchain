<script setup lang="ts">
import { BcDialogConfirm, NotificationsSubscriptionDialog } from '#components'
import type { NotificationSettingsValidatorDashboard, NotificationSettingsAccountDashboard } from '~/types/notifications/subscriptionModal'

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
  is_validator_offline_subscribed: true,
  group_offline_threshold: -20, // means "20% and deselected/unchecked"
  is_attestations_missed_subscribed: true,
  is_block_proposal_subscribed: true,
  is_upcoming_block_proposal_subscribed: false,
  is_sync_subscribed: true,
  is_withdrawal_processed_subscribed: true,
  is_slashed_subscribed: false,
  is_real_time_mode_enabled: false
}

const accountSub: NotificationSettingsAccountDashboard = {
  is_incoming_transactions_subscribed: true,
  is_outgoing_transactions_subscribed: true,
  is_erc20_token_transfers_subscribed: false,
  erc20_token_transfers_threshold: NaN, // means "not in the database yet" (will leave the input field empty with a placeholder)
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
    <Button @click="openSubscriptions({dashboardId:1, validatorSub})">
      Subscribe to notifications for your validators
    </Button>
    <Button @click="openSubscriptions({dashboardId:1, accountSub})">
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
