<script setup lang="ts">
import { BcDialogConfirm, DashboardCreationController } from '#components'

const dialog = useDialog()
const { currentNetwork } = useNetworkStore()

function onClose (answer: boolean) {
  setTimeout(() => {
    alert('response: ' + answer)
  }, 100
  )
}

const openQuestion = (yesLabel?: string, noLabel?: string) => {
  dialog.open(BcDialogConfirm, {
    onClose: response => onClose(response?.data),
    data: {
      title: 'My super question',
      question: 'Are you ready to rumble, or do you have second thoughts?',
      yesLabel,
      noLabel
    }
  })
}

const dashboardCreationControllerModal = ref<typeof DashboardCreationController>()
</script>

<template>
  <div class="container">
    <Button @click="dashboardCreationControllerModal?.show()">
      Create dashboard with free will!
    </Button>
    <Button @click="dashboardCreationControllerModal?.show('validator', currentNetwork)">
      Create validator dashboard with currentNetwork.value forced
    </Button>
    <Button @click="dashboardCreationControllerModal?.show('account')">
      Create account dashboard with account mode forced
    </Button>

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
.container{
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 10px;
  padding: 10px;
}
</style>
