<script setup lang="ts">

const input = ref('123,234,345')
const encrypted = computed(() => toBase64Url(input.value))
const decrypted = computed(() => fromBase64Url(encrypted.value))

</script>

<template>
  <div class="container">
    <div>
      Input <InputText v-model="input" />
    </div>
    <div>
      Base64Url encoded: {{ encrypted }}
    </div>
    <div>
      Base64Url decoded: {{ decrypted }}
    </div>
  </div>
</template>

<style lang="scss" scoped>
.container {
  margin: 10px;
  padding: 10px;
  gap: 10px;
}
</style>
