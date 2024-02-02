<script setup lang="ts">
import type { AdConfiguration } from '~/types/adConfiguration'
const reviveId = '5b200397ccf8a9353bf44ef99b45268c'
interface Props {
  ad: AdConfiguration
}
const props = defineProps<Props>()

const adComponent = ref<HTMLElement | null>(null)
const interval = ref<NodeJS.Timeout | null>(null)

const containerId = computed(() => `${props.ad.key}-${props.ad.jquery_selector}`)

const refreshAd = () => {
  const ins = document.getElementById(containerId.value)?.firstElementChild
  ins?.removeAttribute('data-revive-loaded')
  ins?.removeAttribute('data-revive-seq')
  window.reviveAsync?.[reviveId].refresh()
}
const makeSureReviveIsInitiated = () => {
  const ins = document.getElementById(containerId.value)?.firstElementChild
  if (ins && ins.getAttribute('data-revive-id') && !ins.getAttribute('data-revive-seq') && !ins.getAttribute('data-revive-loaded')) {
    window.reviveAsync?.[reviveId].refresh()
  }
}

onMounted(() => {
  if (!adComponent.value) {
    return
  }
  const target = adComponent.value.parentElement
  if (!target) {
    return
  }

  switch (props.ad.insert_mode) {
    case 'replace':
      target.after(adComponent.value)
      target.remove()
      break
    case 'after':
      target.after(adComponent.value)
      break
    case 'before':
      target.before(adComponent.value)
      break
  }
  if (props.ad.banner_id) {
    if (props.ad.refresh_interval) {
      interval.value = setInterval(refreshAd, props.ad.refresh_interval * 1000)
      setTimeout(makeSureReviveIsInitiated, 1000)
    }
  }
})

onBeforeUnmount(() => {
  if (interval.value) {
    clearInterval(interval.value)
  }
})
</script>
<template>
  <div ref="adComponent">
    <div v-if="ad.banner_id">
      <div class="ad-banner">
        <div :id="containerId" class="revive-container">
          <ins ref="ins" :data-revive-zoneid="ad.banner_id" :data-revive-id="reviveId" />
        </div>
      </div>
    </div>
    <!-- eslint-disable vue/no-v-html -->
    <div v-else v-html="ad.html_content" />
  </div>
</template>

<style lang="scss" scoped>
.ad-banner {
  margin-bottom: var(--padding);
  display: flex;
  justify-content: center;
  align-items: center;

  .revive-container {
    position: relative;
    display: flex;
    justify-content: center;
    align-items: center;
    width: 100%;
    height: 90px;

    :deep(ins iframe) {
      max-width: 100% !important;
    }
  }
}
</style>
