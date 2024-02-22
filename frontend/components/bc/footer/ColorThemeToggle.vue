<script setup lang="ts">
const colorMode = useColorMode()
const switchPosition = ref('dark')

onMounted(() => {
  // intializes the theme with what the user selected the last time or with her/his system preference if (s)he never clicked the button
  switchPosition.value = colorMode.value
  // colorMode.preference should not be read here because its value can be 'system', whereas colorMode.value contains either 'dark' or 'light'
})
</script>

<template>
  <label class="darklight-switch">
    <input
      v-model="switchPosition"
      true-value="light"
      false-value="dark"
      type="checkbox"
      @change="colorMode.preference = switchPosition"
    >
    <span class="slider" />
    <IconColorToggleMoon id="moon" />
    <IconColorToggleSun id="sun" />
  </label>
</template>

<style lang="scss" scoped>
.darklight-switch {
  position: relative;
  display: inline-block;
  width: 36px;
  height: 19px;
  border-radius: 9.5px;
  background-color: #c0adad;
}

.dark-mode .darklight-switch {
  background-color: var(--dark-grey);
}

input {
  opacity: 0;
  width: 0;
  height: 0;
  display: none;
  appearance: none;
  -webkit-appearance: none;
}

.slider {
  position: absolute;
  height: 16px;
  width: 16px;
  left: 1.5px;
  bottom: 1.5px;
  border-radius: 50%;
  background-color: var(--primary-orange);  // do not use --primary-color otherwise the switch becomes invisible in dark mode
  -webkit-transition: 0.2s;
  transition: 0.2s;
}

input:checked+.slider {
  -webkit-transform: translateX(17px);
  -ms-transform: translateX(17px);
  transform: translateX(17px);
}

#moon {
  position: absolute;
  left: 4.5px;
  top: 4.5px;
  color: var(--light-grey-3);
}

#sun {
  position: absolute;
  left: 21.2px;
  top: 4px;
  color: var(--light-grey);
}

.dark-mode .darklight-switch #moon {
  color: var(--light-grey);
}

.dark-mode .darklight-switch #sun {
  color: var(--light-grey-3);
}
</style>
