<script lang="ts" setup>

</script>

<template>
  <BcPageWrapper>
    <div class="settings-container">
      <UserSettingsSubscriptions />
      <UserSettingsEmail />
      <UserSettingsPassword />
      <UserSettingsDeleteAccount />
    </div>
  </BcPageWrapper>
</template>

<style lang="scss" scoped>
.settings-container {
  display: flex;
  flex-direction: column;
  width: 750px;
}
</style>
