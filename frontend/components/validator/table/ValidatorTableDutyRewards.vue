<script setup lang="ts">
import type { ValidatorHistoryDuties } from '~/types/api/common'

const {
  data,
} = defineProps<{
  data: ValidatorHistoryDuties,
}>()

const { t: $t } = useTranslation()

const currencyItems = computed(() => {
  const result = [
    {
      consensusLayerValue: data.attestation_head?.income ?? '0',
    },
    {
      consensusLayerValue: data.attestation_source?.income ?? '0',
    },
    {
      consensusLayerValue: data.attestation_target?.income ?? '0',
    },
    {
      executionLayerValue: data.proposal?.el_income ?? '0',
    },
    {
      consensusLayerValue: data.proposal?.cl_attestation_inclusion_income ?? '0',
    },
    {
      consensusLayerValue: data.proposal?.cl_sync_inclusion_income ?? '0',
    },
    {
      consensusLayerValue: data.proposal?.cl_slashing_inclusion_income ?? '0',
    },
  ]
  return result
})
</script>

<template>
  <BcTooltip
    fit-content
    tooltip-text-align="left"
  >
    <BcFormatAmount
      :currency-items
      has-color
      has-sign-display
      target-unit-crypto="auto"
    />
    <template #tooltip>
      <div>
        <div v-if="data.attestation_head?.income">
          <b>{{ $t('validator.rewards.attestation_head') }}: </b>
          <BcFormatAmount
            :value="data.attestation_head?.income"
            target-unit-crypto="auto"
            has-sign-display
            has-color
          />
        </div>
        <div v-if="data.attestation_source?.income">
          <b>{{ $t('validator.rewards.attestation_source') }}: </b>
          <BcFormatAmount
            :value="data.attestation_source?.income"
            target-unit-crypto="auto"
            has-sign-display
            has-color
          />
        </div>
        <div v-if="data.attestation_target?.income">
          <b>{{ $t('validator.rewards.attestation_target') }}: </b>
          <BcFormatAmount
            :value="data.attestation_target?.income"
            target-unit-crypto="auto"
            has-sign-display
            has-color
          />
        </div>
        <div v-if="data.proposal?.el_income">
          <b>{{ $t('validator.rewards.proposer_el') }}: </b>
          <BcFormatAmount
            :value="data.proposal?.el_income"
            source-currency="elCurrency"
            target-unit-crypto="auto"
            has-sign-display
            has-color
          />
        </div>
        <div v-if="data.proposal?.cl_attestation_inclusion_income">
          <b>{{ $t('validator.rewards.proposer_attestation') }}: </b>
          <BcFormatAmount
            :value="data.proposal?.cl_attestation_inclusion_income"
            target-unit-crypto="auto"
            has-sign-display
            has-color
          />
        </div>
        <div v-if="data.proposal?.cl_sync_inclusion_income">
          <b>{{ $t('validator.rewards.proposer_sync') }}: </b>
          <BcFormatAmount
            :value="data.proposal?.cl_sync_inclusion_income"
            target-unit-crypto="auto"
            has-sign-display
            has-color
          />
        </div>
        <div v-if="data.proposal?.cl_slashing_inclusion_income">
          <b>{{ $t('validator.rewards.proposer_slashing') }}: </b>
          <BcFormatAmount
            :value="data.proposal?.cl_slashing_inclusion_income"
            target-unit-crypto="auto"
            has-sign-display
            has-color
          />
        </div>
      </div>
    </template>
  </BcTooltip>
</template>

<style lang="scss" scoped>
</style>
