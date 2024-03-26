<script setup lang="ts">
import { ResultType, SubCategory, TypeInfo } from '~/types/searchbar'

const props = defineProps<{
  type: ResultType
}>()

const subCategory = computed(() => TypeInfo[props.type].subCategory)

</script>

<template>
  <IconDatatypeTokenIcon v-if="subCategory === SubCategory.Tokens" />
  <IconDatatypeNftIcon v-else-if="subCategory === SubCategory.NFTs" />
  <IconDatatypeEpochIcon v-else-if="subCategory === SubCategory.Epochs" />
  <IconDatatypeBlockIcon v-else-if="subCategory === SubCategory.SlotsAndBlocks" />
  <IconDatatypeTransactionIcon v-else-if="subCategory === SubCategory.Transactions" />
  <IconDatatypeBatchIcon v-else-if="subCategory === SubCategory.Batches" />
  <IconDatatypeContractIcon v-else-if="subCategory === SubCategory.Contracts" />
  <IconDatatypeAccountIcon v-else-if="subCategory === SubCategory.Accounts" />
  <IconDatatypeEnsIcon v-else-if="subCategory === SubCategory.EnsOverview" />
  <IconDatatypeGraffitiIcon v-else-if="subCategory === SubCategory.Graffiti" />
  <IconDatatypeValidatorIcon v-else-if="subCategory === SubCategory.Validators" />
</template>
