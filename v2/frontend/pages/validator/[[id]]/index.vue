<script lang="ts" setup>
import { warn } from 'vue'

const route = useRoute()

warn(`route id: ${route.params.id}, query: ${route.query}`)
</script>

<template>
  <BcPageWrapper>
    <h1>Validator {{ route.params.id }}</h1>
  </BcPageWrapper>
</template>

<style lang="scss">

</style>
