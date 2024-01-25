<script setup lang="ts">
const colorMode = useColorMode()
const switchPosition = ref('dark')

onMounted(() => {
  // intializes the theme with what the user selected the last time or his system preference if (s)he never clicked the button
  switchPosition.value = colorMode.value
  // colorMode.preference should not be read here because its value can be 'system', whereas colorMode.value contains either 'dark' or 'light'
  setTheme()
})

function setTheme () {
  colorMode.preference = switchPosition.value
}
</script>

<template>
  <label class="darklight-switch">
    <input v-model="switchPosition" true-value="light" false-value="dark" type="checkbox" @change="setTheme">
    <span class="slider" />
    <IconColorToggleMoon id="moon" :class="switchPosition === 'dark' ? 'icon-on' : 'icon-off'" />
    <IconColorToggleSun id="sun" :class="switchPosition === 'dark' ? 'icon-off' : 'icon-on'" />
  </label>
</template>

<style lang="scss" scoped>
.darklight-switch {
  position: relative;
  display: inline-block;
  width: 36px;
  height: 19px;
  border-radius: 9.5px;
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
  background-color: var(--primary-color);
  -webkit-transition: .2s;
  transition: .2s;
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
}

#sun {
  position: absolute;
  left: 21.2px;
  top: 4px;
}

.icon-on {
  color: var(--light-grey);
  fill: var(--light-grey);
}

.icon-off {
  color: var(--light-grey-3);
  fill: var(--light-grey-3);
}
</style>
