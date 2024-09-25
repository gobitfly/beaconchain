<script setup lang="ts">
import {
  BcDialogConfirm,
  DashboardCreationController,
  NotificationsManagementSubscriptionDialog,
} from '#components'
import type {
  NotificationSettingsAccountDashboard,
  NotificationSettingsValidatorDashboard,
} from '~/types/api/notifications'

const dialog = useDialog()
const { currentNetwork } = useNetworkStore()

function onClose(answer: boolean) {
  setTimeout(() => {
    alert('response: ' + answer)
  }, 100)
}

const openQuestion = (yesLabel?: string, noLabel?: string) => {
  dialog.open(BcDialogConfirm, {
    data: {
      noLabel,
      question: 'Are you ready to rumble, or do you have second thoughts?',
      title: 'My super question',
      yesLabel,
    },
    onClose: response => onClose(response?.data),
  })
}

const validatorSub: NotificationSettingsValidatorDashboard = {
  group_offline_threshold: 0, // means "deactivated/unchecked"
  is_attestations_missed_subscribed: true,
  is_block_proposal_subscribed: true,
  is_group_offline_subscribed: true,
  is_real_time_mode_enabled: false,
  is_slashed_subscribed: false,
  is_sync_subscribed: true,
  is_upcoming_block_proposal_subscribed: false,
  is_validator_offline_subscribed: true,
  is_webhook_discord_enabled: true,
  is_withdrawal_processed_subscribed: true,
  webhook_url: 'http://bablabla',
}

const accountSub: NotificationSettingsAccountDashboard = {
  erc20_token_transfers_value_threshold: 0,
  is_erc20_token_transfers_subscribed: false,
  is_erc721_token_transfers_subscribed: true,
  is_erc1155_token_transfers_subscribed: false,
  is_ignore_spam_transactions_enabled: true,
  is_incoming_transactions_subscribed: true,
  is_outgoing_transactions_subscribed: true,
  is_webhook_discord_enabled: true,
  subscribed_chain_ids: [ 17000 ],
  webhook_url: 'http://bablabla',
}

function openSubscriptions(props: any) {
  dialog.open(NotificationsManagementSubscriptionDialog, {
    data: props,
    onClose: response => onClose(response?.data),
  })
}

const dashboardCreationControllerModal
  = ref<typeof DashboardCreationController>()
</script>

<template>
  <div class="container">
    <Button @click="dashboardCreationControllerModal?.show()">
      Create dashboard with free will!
    </Button>
    <Button
      @click="
        dashboardCreationControllerModal?.show('validator', currentNetwork)
      "
    >
      Create validator dashboard with currentNetwork.value forced
    </Button>
    <Button @click="dashboardCreationControllerModal?.show('account')">
      Create account dashboard with account mode forced
    </Button>
    Note: to test the saving of the options to the API, open the dialogs from
    the notification dashboard.<br>The communication with the API is
    implemented there.
    <Button
      @click="
        openSubscriptions({
          dashboardType: 'validator',
          initialSettings: validatorSub,
          saveUserSettings: () => {},
        })
      "
    >
      Subscribe to notifications for your validators
    </Button>
    <Button
      @click="
        openSubscriptions({
          dashboardType: 'account',
          initialSettings: accountSub,
          saveUserSettings: () => {},
        })
      "
    >
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

    <DashboardCreationController
      ref="dashboardCreationControllerModal"
      class="modal-controller"
      :display-mode="'modal'"
    />
  </div>
</template>

<style lang="scss" scoped>
.container {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 10px;
  padding: 10px;
}
</style>
