<script setup lang="ts">
import { BcDialogConfirm, NotificationsSubscriptionDialog } from '#components'
import type { NotificationEventsValidatorDashboard, NotificationEventsAccountDashboard } from '~/types/notifications/subscriptionModal'

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

const validatorSub: NotificationEventsValidatorDashboard = {
  validator_offline: true,
  group_offline: -20, // means "20% and deselected/unchecked"
  attestations_missed: true,
  block_proposal: true,
  upcoming_block_proposal: false,
  sync: true,
  withdrawal_processed: true,
  slashed: false,
  realtime_mode: false
}

const accountSub: NotificationEventsAccountDashboard = {
  incoming_transactions: true,
  outgoing_transactions: true,
  track_erc20_token_transfers: null, // means "not in the database yet" (will leave the input field empty with a placeholder)
  track_erc721_token_transfers: true,
  track_erc1155_token_transfers: false,
  networks: [17000],
  ignore_spam_transactions: true
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
    <Button @click="openSubscriptions({validatorSub})">
      Subscribe to notifications for your validators
    </Button>
    <Button @click="openSubscriptions({accountSub})">
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
