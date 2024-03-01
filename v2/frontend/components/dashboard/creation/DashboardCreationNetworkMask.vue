<script lang="ts" setup>

</script>

<template>
  <div class="mask_container">
    <div class="element_container">
      <div class="big_text">
        Add a new dashboard
      </div>
      <div class="subtitle_text">
        What network are your validators on?
      </div>
      <div class="network_button_container">
        <Button>Ethereum</Button>
        <Button>Gnosis</Button>
      </div>
      <div class="continue_button_container">
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

      .network_button_container{
        display: flex;
        gap: 10px;
      }

      .continue_button_container{
        display: flex;
        justify-content: flex-end;
      }
    }
  }
</style>
