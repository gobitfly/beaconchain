<script setup lang="ts">
const year = new Date().getFullYear()
const version = useRuntimeConfig().public.gitVersion
const { t: $t } = useTranslation()
const v1Domain = useV1Domain()
</script>

<template>
  <div class="main-footer">
    <nav class="main-footer__links">
      <section>
        <BcText
          is="h4"
          variant="md"
          class="main-footer__headline"
        >
          {{ $t('footer.legal_notices') }}
        </BcText>
        <ul class="main-footer__list">
          <li>
            <BcLink
              :to="`${v1Domain}/imprint`"
              class="link"
            >
              {{ $t('footer.imprint') }}
            </BcLink>
          </li>
          <li>
            <BcLink
              :to="LINK.termsOfService"
              class="link"
              external
            >
              {{ $t('footer.terms') }}
            </BcLink>
          </li>
          <li>
            <BcLink
              :to="LINK.privacy"
              class="link"
              external
            >
              {{ $t('footer.privacy') }}
            </BcLink>
          </li>
        </ul>
      </section>
      <section
        class="main-footer__section"
      >
        <BcText
          is="h4"
          variant="md"
        >
          {{ $t('footer.resources') }}
        </BcText>
        <ul class="main-footer__list">
          <li>
            <BcLink
              class="link"
              to="/pricing"
            >
              {{ $t('footer.premium') }}
            </BcLink>
          </li>
          <li>
            <BcLink
              class="link"
              :to="LINK.shop"
            >
              {{ $t('footer.shop') }}
            </BcLink>
          </li>
          <li>
            <BcLink
              class="link"
              :to="LINK.status"
              external
            >
              {{ $t('footer.status') }}
            </BcLink>
          </li>
        </ul>
      </section>
      <section
        class="main-footer__section"
      >
        <BcText
          is="h4"
          variant="md"
        >
          {{ $t('footer.links') }}
        </BcText>
        <ul class="main-footer__list">
          <li>
            <BcLink
              class="link"
              :to="LINK.discord"
            >
              {{ $t('footer.discord') }}
            </BcLink>
          </li>
          <li>
            <BcLink
              class="link"
              :to="LINK.x"
            >
              {{ $t('footer.x') }}
            </BcLink>
          </li>
          <li>
            <BcLink
              class="link"
              :to="LINK.github"
            >
              {{ $t('footer.github') }}
            </BcLink>
          </li>
          <li>
            <BcLink
              class="link"
              :to="LINK.githubMobileApp"
            >
              {{ $t('footer.github_mobile_app') }}
            </BcLink>
          </li>
        </ul>
      </section>
    </nav>
    <p class="main-footer__meta">
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
    </p>
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
  padding: var(--padding-large);
  margin-top: 64px;
  background-color: var(--container-background);
  align-items: center;

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
    flex-direction: column;
    gap: var(--padding-large);
    max-width: $breakpoint-sm;

    @media (min-width: $breakpoint-md) {
      flex-direction: row;
      justify-content: space-between;
    }
  }

  .main-footer__list {
    margin-top: var(--padding-medium);
    list-style: none;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: var(--padding);
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
