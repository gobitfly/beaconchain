<script lang="ts" setup>
const { isLoggedIn } = useUserStore()

useBcSeo('user_settings.title')

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
  position: relative;
  margin-left: auto;
  margin-right: auto;
  width: 750px;

  display: flex;
  flex-direction: column;
  gap: var(--padding);
}
</style>
