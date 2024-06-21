<script lang="ts" setup>
const { isLoggedIn } = useUserStore()

useBcSeo('user_settings.title')

const buttonsDisabled = ref(false)

// TODO: This seems to be triggered to soon
if (!isLoggedIn.value) {
  // only users that are logged in can view this page
  await navigateTo('/login')
}
</script>

<template>
  <BcPageWrapper>
    <div class="settings-container">
      <UserSettingsSubscriptions v-model="buttonsDisabled" />
      <UserSettingsEmail v-model="buttonsDisabled" />
      <UserSettingsPassword v-model="buttonsDisabled" />
      <UserSettingsDeleteAccount v-model="buttonsDisabled" />
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
