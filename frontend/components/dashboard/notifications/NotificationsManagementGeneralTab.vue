<script lang="ts" setup>
const doNotDisturbToggle = ref(false)
const emailToggle = ref(false)
const pushToggle = ref(false)
</script>

<template>
  <div class="container">
    <div class="row divider">
      <div>
        <span>Do not disturb</span>
        <span class="explanation">Mutes all notifications</span>
      </div>
      <BcToggle v-model="doNotDisturbToggle" />
    </div>
    <div class="row">
      <div>
        E-Mail Notifications
      </div>
      <BcToggle v-model="emailToggle" />
    </div>
    <div class="row">
      <div>
        Push Notifications
      </div>
      <BcToggle v-model="pushToggle" />
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/fonts.scss";

.container{
  border: unset;
  margin-top: var(--padding-xl);
  display: flex;
  flex-direction: column;
  align-items: center;
  @include fonts.small_text_bold;

  .row {
    width: 100%;
    max-width: 500px;
    padding: var(--padding);
    display: flex;
    justify-content: space-between;

    .explanation{
      @include fonts.tiny_text;
      color: var(--text-color-discreet);
      margin-left: var(--padding-small);
    }

    &.divider {
      padding-bottom: calc(var(--padding-large) - var(--padding-small));
      margin-bottom: var(--padding-small);
      border-bottom: 1px solid var(--container-border-color);
    }
  }
}
</style>
