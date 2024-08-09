<script setup lang="ts">
const toast = useBcToast()

const openToast = (
  severity: 'error' | 'info' | 'success',
  summary: string,
  detail: string,
  group?: string,
) => {
  switch (severity) {
    case 'error':
      toast.showError({
        detail,
        group,
        summary,
      })
      break
    case 'info':
      toast.showInfo({
        detail,
        group,
        summary,
      })
      break
    case 'success':
      toast.showSuccess({
        detail,
        group,
        summary,
      })
      break
  }
}
</script>

<template>
  <div class="container">
    <Button
      @click="
        openToast(
          'success',
          'Create Group',
          'You succcessfully created group Haudrauf',
        )
      "
    >
      trigger success
    </Button>
    <Button
      @click="
        openToast(
          'info',
          'You are happy',
          'And dont let anyone else tell you differently!',
        )
      "
    >
      trigger info
    </Button>
    <Button
      @click="
        openToast(
          'error',
          'Webservice Error',
          'This is not the page you are looking for!',
          '404',
        )
      "
    >
      trigger error code 401
    </Button>
    <Button
      @click="
        openToast(
          'error',
          'Webservice Error',
          'GET: https://thisismydomain.com/api/data',
          '501',
        )
      "
    >
      trigger error code 501
    </Button>
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
