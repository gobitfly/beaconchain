<script setup lang="ts">
const colorMode = useColorMode()
const switchPosition = ref('dark')

onMounted(() => {
  // intializes the theme with what the user selected the last time or with
  // her/his system preference if (s)he never clicked the button
  switchPosition.value = colorMode.value
  // colorMode.preference should not be read here because its value can be 'system',
  // whereas colorMode.value contains either 'dark' or 'light'
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
    <BcIcon
      id="moon"
      size="sm"
      name="moon"
    />
    <BcIcon
      id="sun"
      size="sm"
      name="sun"
    />
  </label>
</template>

<style lang="scss" scoped>
.darklight-switch {
  position: relative;
  display: inline-block;
  width: 2.25rem;
  height: 1.1875rem;
  border-radius: .5938rem;
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
  height: 1rem;
  width: 1rem;
  left: .0938rem;
  bottom: .0938rem;
  border-radius: 50%;
  background-color: var(--primary-color);
  -webkit-transition: 0.2s;
  transition: 0.2s;
}

input:checked + .slider {
  transform: translateX(1.0625rem);
}

#moon {
  position: absolute;
  left: .2188rem;
  top: .2188rem;
  color: var(--light-grey-3);
}

#sun {
  position: absolute;
  right: .1875rem;
  translate: -0.025rem -0.025rem;
  top: .25rem;
  color: var(--light-grey);
}

.dark-mode .darklight-switch #moon {
  color: var(--light-grey);
}

.dark-mode .darklight-switch #sun {
  color: var(--light-grey-3);
}
</style>
