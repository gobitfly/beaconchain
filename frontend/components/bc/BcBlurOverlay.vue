<template>
  <div class="gradient-blur" />
</template>
<style lang="scss" scoped>
.gradient-blur {
  background: linear-gradient(rgba(233, 233, 233, 0), rgba(233, 233, 233, 1));
  position: absolute;
  z-index: 5;
  height: 75%;
}

.dark-mode {
  .gradient-blur {
    background: linear-gradient(rgba(35, 32, 36, 0), rgba(35, 32, 36, 1));
  }
}
</style>
