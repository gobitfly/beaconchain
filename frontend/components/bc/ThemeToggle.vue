<script setup lang="ts">
const colorMode = useColorMode()

function toggleTheme() {
  colorMode.preference = colorMode.preference === 'dark' ? 'light' : 'dark'
}
</script>

<template>
  <Button :label="`toggle: ${colorMode.value}`" size="small" @click="toggleTheme" />

  <label class="darklight-switch">
    <input type="checkbox">
    <span class="slider" />
  </label>
</template>

<style lang="scss" scoped>
.darklight-switch {
  position: relative;
  display: inline-block;
  width: 36px;
  height: 19px;
}

.switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.slider {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  border-radius: 16px;
  background-color: var(--dark-grey);
  -webkit-transition: .2s;
  transition: .2s;
}

.slider:before {
  position: absolute;
  content: "";
  height: 16px;
  width: 16px;
  left: 1.5px;
  bottom: 1.5px;
  border-radius: 50%;
  background-color: var(--primary-color);
  -webkit-transition: .2s;
  transition: .2s;
}

input:checked+.slider:before {
  -webkit-transform: translateX(17px);
  -ms-transform: translateX(17px);
  transform: translateX(17px);
}
</style>
