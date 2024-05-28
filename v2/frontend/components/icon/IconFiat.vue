<script setup lang="ts">
import type { FiatCurrency } from '~/types/currencies'

interface Props {
  currency?: FiatCurrency
}
defineProps<Props>()
</script>
<template>
  <img :src="`../img/flags/${currency}.svg`">
</template>
