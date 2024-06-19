<script lang="ts" setup>
const { isLoggedIn } = useUserStore()

if (!isLoggedIn.value) {
  // only users that are logged in can view this page
  await navigateTo('../login')
}
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
  width: 750px;

  display: flex;
  flex-direction: column;
  gap: var(--padding)
}
</style>
