<script setup lang="ts">
const year = new Date().getFullYear()
const version = useRuntimeConfig().public.gitVersion
</script>

<template>
  <div class="main-footer">
    <div class="main-footer__links">
      <BcFooterLinkList />
    </div>
    <div class="main-footer__meta">
      <span>
        © bitfly explorer GmbH {{ year }}
      </span>
      <span class="main-footer__meta_seperator">
        |
      </span>
      <span>
        v{{ version }}
      </span>
      <span class="main-footer__meta_seperator">
        |
      </span>
      <BcThemeToggle />
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use '~/assets/css/breakpoints' as *;

.main-footer {
  display: flex;
  position: relative;
  box-sizing: border-box;
  width: 100%;
  flex-direction: column;
  gap: var(--padding-large);
  padding-block: var(--padding-large);
  margin-top: 64px;
  background-color: var(--container-background);

  @media (min-width: 600px) {
    // large screen
    justify-content: center;
  }

  @media (max-width: 599.9px) {
    // mobile
    justify-content: flex-start;
    padding-left: var(--padding-large);
  }

  .main-footer__links {
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

  .main-footer__meta {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--padding);
    .main-footer__meta_seperator {
      display: none;
    }

    @media (min-width: $breakpoint-sm) {
      flex-direction: row;
      justify-content: center;
      align-items: center;
      .main-footer__meta_seperator {
        display: inline;
      }
    }
  }
}
</style>
