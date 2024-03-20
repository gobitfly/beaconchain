<script setup lang="ts">
import { BcDialogConfirm } from '#components'

const dialog = useDialog()

function onClose (answer: boolean) {
  setTimeout(() => {
    alert('response: ' + answer)
  }, 100
  )
}

const openQuestion = (yesLabel?: string, noLabel?: string) => {
  dialog.open(BcDialogConfirm, {
    props: {
      header: 'My super question'
    },
    onClose: response => onClose(response?.data),
    data: {
      question: 'Are you ready to rumble, or do you have second thoughts?',
      yesLabel,
      noLabel
    }
  })
}
</script>

<template>
  <div class="container">
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
