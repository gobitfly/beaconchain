<script setup lang="ts">
const input = ref('12,32,123,1,35,4')
const encrypted = computed(() => toBase64Url(input.value))
const decrypted = computed(() => fromBase64Url(encrypted.value))

const input2 = ref('MTIsMzIsMTIzLDEsMzUsNA')
const decrypted2 = computed(() => fromBase64Url(input2.value))
const encrypted2 = computed(() => toBase64Url(decrypted2.value))
</script>

<template>
  <div class="container">
    <h>Encoding</h>
    <div>Input <InputText v-model="input" /></div>
    <div>Base64Url encoded: {{ encrypted }}</div>
    <div>Base64Url decoded: {{ decrypted }}</div>
  </div>
  <div class="container">
    <h>Reverse Encoding</h>
    <div>Base64Url Input <InputText v-model="input2" /></div>
    <div>Decrypted: {{ decrypted2 }}</div>
    <div>Base64Url decoded: {{ encrypted2 }}</div>
  </div>
</template>

<style lang="scss" scoped>
.container {
  margin: 10px;
  padding: 10px;
  gap: 10px;
}
</style>
