<script lang="ts" setup>
// TODO: Styling
// TODO: Logic
const visible = defineModel<boolean>()

</script>

<template>
  <BcDialog
    v-model="visible"
    header="Paired Devices"
    class="paired-devices-modal-container"
  >
    <Button class="done-button" :label="$t('navigation.done')" @click="visible = false" />
  </BcDialog>
</template>

<style lang="scss" scoped>
:global(.paired-devices-modal-container) {
  max-width: 790px;
}

.done-button {
  position: absolute;
  bottom: calc(var(--padding-large) + var(--padding));
  right: calc(var(--padding-large) + var(--padding));
}
</style>
