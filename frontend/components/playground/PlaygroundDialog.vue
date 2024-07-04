<script setup lang="ts">
import { BcDialogConfirm, NotificationsSubscriptionDialog } from '#components'
import type { ValidatorSubscriptionState, AccountSubscriptionState } from '~/types/subscriptionModal'

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

const validatorSub: ValidatorSubscriptionState = {
  offlineValidator: true,
  offlineGroup: 40,
  missedAttestations: true,
  proposedBlock: true,
  upcomingProposal: false,
  syncCommittee: true,
  withdrawn: true,
  slashed: false,
  realTime: false
}

const accountSub: AccountSubscriptionState = {
  incoming: true,
  outgoing: true,
  erc20: -1,
  erc721: true,
  erc1155: false,
  networks: [17000, 10200],
  ignoreSpam: true
}

function openSubscriptions (props: any) {
  dialog.open(NotificationsSubscriptionDialog, {
    data: props,
    onClose: modified => onClose(modified?.data)
  })
}
</script>

<template>
  <div class="container">
    <Button @click="openSubscriptions({validatorSub, premiumUser: true})">
      Subscribe to notifications for your validators
    </Button>
    <Button @click="openSubscriptions({accountSub, premiumUser: true})">
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
