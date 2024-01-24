<script setup lang="ts">
const colorMode = useColorMode()

const toggleTheme = () => {
  colorMode.preference = colorMode.preference === 'dark' ? 'light' : 'dark'
}

</script>
<template>
  <Button :label="`toggle: ${colorMode.value}`" size="small" @click="toggleTheme" />
</template>

<style lang="scss" scoped>
</style>
