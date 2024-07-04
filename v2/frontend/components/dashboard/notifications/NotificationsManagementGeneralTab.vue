<script lang="ts" setup>

</script>

<template>
  <div class="container">
    <div class="row divider">
      <div>
        E-Mail Notifications
      </div>
      <div>
        Switch
      </div>
    </div>
    <div class="row">
      <div>
        Push Notifications
      </div>
      <div>
        Switch
      </div>
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

    &.divider {
      padding-bottom: calc(var(--padding-large) - var(--padding-small));
      margin-bottom: var(--padding-small);
      border-bottom: 1px solid var(--container-border-color);
    }
  }
}
</style>
