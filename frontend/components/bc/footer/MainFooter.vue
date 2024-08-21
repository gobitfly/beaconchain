<script setup lang="ts"></script>

<template>
  <div class="main-footer">
    <div class="links">
      <BcFooterLinkList />
    </div>
    <div class="very-last-line">
      <BcFooterCopyrightAndVersion />&nbsp;|&nbsp;<BcFooterColorThemeToggle />
    </div>
  </div>
</template>

<style lang="scss" scoped>
.main-footer {
  display: flex;
  position: relative;
  box-sizing: border-box;
  width: 100%;
  flex-direction: column;
  gap: var(--padding-large);
  padding-top: var(--padding-large);
  padding-bottom: var(--padding-large);
  padding-left: var(--padding);
  padding-right: var(--padding);
  margin-top: 64px;
  background-color: var(--container-background);

  @media (min-width: 600px) {
    // large screen
    justify-content: center;
  }

  @media (max-width: 599.9px) {
    // mobile
    justify-content: flex-start;
  }

  .links {
    display: flex;
    position: relative;

    @media (min-width: 600px) {
      // large screen
      justify-content: center;
      flex-direction: row;
      gap: 130px;
    }

    @media (max-width: 599.9px) {
      // mobile
      justify-content: flex-start;
      align-items: flex-start;
      flex-direction: column;
    }
  }

  .very-last-line {
    display: flex;
    position: relative;
    flex-direction: row;
    padding-bottom: 13px;

    @media (min-width: 600px) {
      // large screen
      justify-content: center;
    }

    @media (max-width: 599.9px) {
      // mobile
      justify-content: flex-start;
    }
  }
}
</style>
