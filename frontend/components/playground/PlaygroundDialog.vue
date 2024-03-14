<script setup lang="ts">
import { BcDialogConfirm } from '#components'

const dialog = useDialog()

function onClose (answer: boolean) {
  setTimeout(() => {
    alert('response: ' + answer)
  }, 100
  )
}

const openQuestion = () => {
  dialog.open(BcDialogConfirm, {
    props: {
      header: 'My super question'
    },
    onClose: response => onClose(response?.data),
    data: {
      question: 'Are you ready to rumble?'
    }
  })
}
</script>

<template>
  <Button @click="openQuestion">
    Open Question
  </Button>
</template>

<style lang="scss" scoped>
</style>
