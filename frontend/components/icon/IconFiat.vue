<script setup lang="ts">
interface Props {
  currency?: CurrencyCodeFiat,
}
defineProps<Props>()
</script>

<template>
  <img :src="`../img/flags/${currency}.svg`">
</template>
