<script lang="ts" setup>

</script>

<template>
  <div class="mask_container">
    <div class="element_container">
      <div class="big_text">
        Add a new dashboard
      </div>
      <div class="subtitle_text">
        What would you like to monitor?
      </div>
      <div class="row_container">
        <Button>Accounts</Button>
        <Button>Validators</Button>
      </div>
      <div class="row_container">
        <InputText />
        <Button>Continue</Button>
      </div>
    </div>
  </div>
</template>

<style lang="scss">
  .mask_container{
    width: 460px;
    height: 248px;

    .element_container{
      display: flex;
      flex-direction: column;
      gap: 10px;

      .row_container{
        display: flex;
        gap: 10px;
      }
    }
  }
</style>
