<script setup lang="ts">
import { ResultType, SubCategory, TypeInfo } from '~/types/searchbar'

const props = defineProps({
  type: { type: String, required: true }
})

const subCategory = TypeInfo[props.type as ResultType].subCategory

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
  <IconDatatypeEnsIcon v-else-if="subCategory === SubCategory.EnsSystem" />
  <IconDatatypeGraffitiIcon v-else-if="subCategory === SubCategory.Graffiti" />
  <IconDatatypeValidatorIcon v-else-if="subCategory === SubCategory.Validators" />
</template>
