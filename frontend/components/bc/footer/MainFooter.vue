<script setup lang="ts">

</script>

<template>
  <div id="main-footer">
    <div id="links">
      <BcFooterLinkList />
    </div>
    <div id="very-last-line">
      <BcFooterCopyrightAndVersion />&nbsp;|&nbsp;<BcFooterColorThemeToggle />
    </div>
  </div>
</template>

<style lang="scss" scoped>
#main-footer {
  display: flex;
  flex-direction: column;
  left: 0;
  right: 0;
  margin-top: 4px;
  margin-bottom: 13px;
  margin-left: 10px;
  @media (min-width: 600px) { // large screen
    justify-content: center;
  }
  @media (max-width: 600px) { // mobile
    justify-content: flex-start;
    width: 100%;
  }
}

#links {
  display: flex;
  @media (min-width: 600px) { // large screen
    justify-content: center;
    flex-direction: row;
    gap: 130px;
  }
  @media (max-width: 600px) { // mobile
    justify-content: flex-start;
    align-items: flex-start;
    flex-direction: column;
    gap: 0px;
  }
}

#very-last-line {
  display: flex;
  flex-direction: row;
  margin-top: 10px;
  @media (min-width: 600px) { // large screen
    justify-content: center;
  }
  @media (max-width: 600px) { // mobile
    justify-content: flex-start;
  }
}
</style>
