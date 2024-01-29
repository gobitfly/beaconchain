<script setup lang="ts">

</script>

<template>
  <div id="main-footer">
    <div id="columns-of-links">
      <BcFooterColumnsOfLinks />
    </div>
    <div id="very-last-line">
      <BcFooterCopyrightAndVersion />&nbsp;|&nbsp;<BcFooterColorThemeToggle />
    </div>
  </div>
</template>

<style lang="scss" scoped>
#main-footer {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  justify-content: center;
  align-items: center;
  margin-top: 4px;
  margin-bottom: 13px;
}

#columns-of-links {
  display: flex;
  gap: 130px;
  justify-content: center;
  margin-bottom: 10px;
}

#very-last-line {
  justify-content: center;
  display: flex;
}
</style>
