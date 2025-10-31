<script setup lang="ts">
const input1 = ref('My name')
const input2 = ref('Not empty')
const input3 = ref('')
</script>

<template>
  <div class="container">
    <div>
      <BcInputLabel
        :value="input1"
        @set-value="(v: string) => (input1 = v)"
      />
    </div>
    <div>
      <BcInputLabel
        :value="input2"
        @set-value="(v: string) => (input2 = v)"
      />
    </div>
    <div>
      <BcInputLabel
        :value="input3"
        :label="input3 || 'Default'"
        :can-be-empty="true"
        @set-value="(v: string) => (input3 = v)"
      />
    </div>
    <div>
      <BcInputLabel
        :value="input1"
        :disabled="true"
      />
    </div>
  </div>
</template>

<style lang="scss" scoped>
.container {
  margin: 10px;
  padding: 10px;
  gap: 10px;
  display: flex;
  overflow: auto;
  background: var(--container-background);
  color: var(--container-color);
  border: solid 1px var(--container-border-color);

  > div {
    width: 200px;
    border: solid 1px var(--container-border-color);
    padding: var(--padding-small);
  }
}
</style>
