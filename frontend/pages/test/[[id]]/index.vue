<script setup lang="ts">
const route = useRoute()
const router = useRouter()
const onClick = () => {
  router.replace('/test/1')
}
</script>

<template>
  <div>
    Page: test
    <div>
      route: {{ route.fullPath }}
    </div>
    <div>
      <button
        @click="onClick"
      >
        click
      </button>
    </div>
  </div>
</template>

<style scoped></style>
