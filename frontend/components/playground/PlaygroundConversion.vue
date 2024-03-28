<script setup lang="ts">
import type { Currency } from '~/types/currencies'

const { setCurrency, currency } = useCurrency()
const { latestState } = useLatestStateStore()

const onCurrencyChange = (event: Event) => {
  const select = event.target as HTMLSelectElement
  if (select?.value) {
    setCurrency(select?.value as Currency)
  }
}

</script>
<template>
  <div>
    Conversions
    <select :value="currency" @change="onCurrencyChange($event)">
      <option value="NAT">
        Native
      </option>
      <option value="ETH">
        Ethereum
      </option>
      <option value="EUR">
        Euro
      </option>
      <option value="GNO">
        Gnosis
      </option>
    </select>
    <div>
      1Eth exact :<BcFormatValue value="1000000000000000000" />
    </div>
    <div>
      1Eth+ (custom tooltip):<BcFormatValue value="1000000010000000000">
        <template #tooltip="{data:{label, tooltip}}">
          Dynamic tooltip: {{ label }} -> {{ tooltip }}
        </template>
      </BcFormatValue>
    </div>
    <div>
      less than 1Eth+ :<BcFormatValue value="0000000010002000001" />
    </div>
    <div>
      less than Wei+ :<BcFormatValue value="0000000000000001000" />
    </div>
    <div>
      less than 1Eth in ETH :<BcFormatValue value="0000000010002000001" :options="{minUnit:'MAIN'}" />
    </div>

    <div>
      1 GNO :<BcFormatValue value="1000000000000000000" :options="{sourceCurrency: 'GNO'}" />
    </div>

    <div>
      1 ETH to GNO :<BcFormatValue value="1000000000000000000" :options="{targetCurrency: 'GNO'}" />
    </div>
    <div>
      1 GNO to GNO :<BcFormatValue value="1000000000000000000" :options="{sourceCurrency: 'GNO', targetCurrency: 'GNO'}" />
    </div>
    <div>
      1 GNO to EUR :<BcFormatValue value="1000000000000000000" :options="{sourceCurrency: 'GNO', targetCurrency: 'EUR'}" />
    </div>
    <div>
      1Eth exact, fixed 0 decimals :<BcFormatValue value="1000000000000000000" :options="{fixedDecimalCount: 0}" />
    </div>
    <div>
      1Eth+ fixed 2 decimals:<BcFormatValue value="1001000000000000000" :options="{fixedDecimalCount: 2}" />
    </div>

    <div>
      1Eth+ with Plus Sign:<BcFormatValue value="1001000000000000000" :options="{addPlus: true}" />
    </div>

    <div>
      -value:<BcFormatValue value="-10010000000000" :options="{addPlus: true}" />
    </div>

    <div>
      Positive with color:<BcFormatValue value="1001000000000000000" :options="{addPlus: true}" :use-colors="true" />
    </div>

    <div>
      Negative with color:<BcFormatValue value="-10010000000000" :options="{addPlus: true}" :use-colors="true" />
    </div>
    <div>
      Negative with custom color:<BcFormatValue value="-10010000000000" :options="{addPlus: true}" :use-colors="true" negative-class="bad-color" />
    </div>
    <div>
      Input in  gwei [1001000]: <BcFormatValue value="1001000" :options="{sourceUnit:'GWEI'}" />
    </div>
    <div>
      Input in eth [2]: <BcFormatValue value="2" :options="{sourceUnit:'MAIN'}" />
    </div>
    <div>
      Input in eth [2] fixed out in GWEI: <BcFormatValue value="2" :options="{sourceUnit:'MAIN', fixedUnit:'GWEI'}" />
    </div>
    <div>
      Input in eth [2] fixed out in WEI: <BcFormatValue value="2" :options="{sourceUnit:'MAIN', fixedUnit:'WEI'}" />
    </div>
  </div>
  <b>
    Format numbers
  </b>
  <div>100000, no settings: <BcFormatNumber :value="100000" /></div>
  <div>100000.1234, no settings: <BcFormatNumber :value="100000.1234" /></div>
  <div>100000, min 2 decimals: <BcFormatNumber :value="100000" :min-decimals="2" /></div>
  <div>100000.1234, min/max 3: <BcFormatNumber :value="100000.1234" :min-decimals="3" :max-decimals="3" /></div>
  <div>0, no settings: <BcFormatNumber :value="0" /></div>
  <div>0.00001, no settings: <BcFormatNumber :value="0.00001" /></div>
  <div>0.01, no settings: <BcFormatNumber :value="0.01" /></div>
  <div>no value, no settings: <BcFormatNumber /></div>
  <div>no value, default '-': <BcFormatNumber default="-" /></div>
  <div>-100000, no settings: <BcFormatNumber :value="-100000" /></div>

  <b>
    Format percent
  </b>

  <div>
    1234567.89123, color, +:
    <BcFormatPercent :percent="1234567.89123" :color-break-point="80" :add-positive-sign="true" />
  </div>
  <div>
    -1234567.89123, color, +:
    <BcFormatPercent :percent="-1234567.89123" :color-break-point="80" :add-positive-sign="true" />
  </div>

  <div>
    1 - no settings
    <BcFormatPercent :percent="1" />
  </div>
  <b>
    Format Epochs time passed
  </b>
  <div>
    Epoch 1 ->
    <BcFormatTimePassed :value="1" />
  </div>

  <div>
    Epoch 272684 ->
    <BcFormatTimePassed :value="272684" />
  </div>
  <div>
    latest Epoch ->
    <BcFormatTimePassed :value="latestState?.currentEpoch" />
  </div>
  <div>
    latest Epoch - 1 ->
    <BcFormatTimePassed :value="(latestState?.currentEpoch ?? 1) - 1" />
  </div>
  <div>
    latest Epoch - 10 ->
    <BcFormatTimePassed :value="(latestState?.currentEpoch ?? 10) - 10" />
  </div>
  <div>
    next Epoch ->
    <BcFormatTimePassed :value="(latestState?.currentEpoch ?? 0 )+ 1" />
  </div>
  <div>
    next Epoch no tick ->
    <BcFormatTimePassed :value="(latestState?.currentEpoch ?? 0 )+ 1" :no-update="true" />
  </div>
  <div>
    the Epoch after ->
    <BcFormatTimePassed :value="(latestState?.currentEpoch ?? 0) + 2" />
  </div>
  <div>
    latest Epoch long format->
    <BcFormatTimePassed :value="latestState?.currentEpoch" unit-length="long" />
  </div>
  <div>
    latest Epoch short format->
    <BcFormatTimePassed :value="latestState?.currentEpoch" unit-length="short" />
  </div>
</template>

<style lang="scss" scoped>

:deep(.bad-color){
  color: pink
}
</style>
