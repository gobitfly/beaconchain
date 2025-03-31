<script setup lang="ts">
import BcIcon from '~/components/bc/icon/BcIcon.vue'
import type { InternalPostSearchResponse } from '~/types/api/search'

const props = defineProps<{
  hasPremiumPerkBulkAdding: boolean,
  hasReachedLimit: boolean,
  isGuestDashboard: boolean,
}>()

const { fetch } = useCustomFetch()
const { t: $t } = useTranslation()

const input = defineModel<string>()

const validatorDashboardOverviewStore = useValidatorDashboardOverviewStore()
const {
  overview,
} = storeToRefs(validatorDashboardOverviewStore)
const { chainIdByDefault } = useRuntimeConfig().public

const currentDashboardNetwork = computed(() => overview.value?.network ?? chainIdByDefault)
const {
  data,
  error,
  execute,
  status,
} = useAsyncData(
  'validator_search',
  () => fetch<InternalPostSearchResponse>('SEARCH', {
    body: {
      input: input.value,
      networks: [ currentDashboardNetwork.value ],
    },
  }), {
    immediate: false,
  },
)
const hasError = computed(() => !!error.value)
const results = computed(() => data.value?.data)
const isLoading = computed(() => status.value === 'pending')

const handleSearch = (input: string) => {
  if (!input.length) return
  error.value = null
  execute()
}
const emit = defineEmits<{
  (e: 'submit', value: InternalPostSearchResponse['data'][number] | undefined): void,
}>()
const handleSubmit = (result: InternalPostSearchResponse['data'][number] | undefined) => {
  emit('submit', result)
}
const isDisabled = (type: InternalPostSearchResponse['data'][number]['type']) => {
  if (props.hasReachedLimit) {
    return true
  }
  if (
    !props.hasPremiumPerkBulkAdding
    && (type !== 'validator' && type !== 'validator_list')
  ) {
    return true
  }
  if (
    props.isGuestDashboard
    && ((type !== 'validator' && type !== 'validator_list'))
  ) {
    return true
  }
  return false
}
</script>

<template>
  <BcInputSearch
    v-model="input"
    :results
    :is-loading
    should-select-first-result
    :should-clear-on-submit="false"
    has-focus
    :has-error
    @search="handleSearch"
    @submit="handleSubmit"
  >
    <template #empty>
      {{ $t('dashboard.validator.management.search.empty') }}
    </template>
    <template #result="{ item }">
      <div
        v-if="item.type === 'validator'"
        class="dashboard-validator-management-modal-search__item"
        :class="{ 'dashboard-validator-management-modal-search__item--disabled': isDisabled(item.type) }"
      >
        <BcIcon
          name="desktop"
        />
        <span class="dashboard-validator-management-modal-search__item-validator_info">
          1 {{ $t('common.validator', 1) }}
          <BcIcon
            v-if="isDisabled(item.type)"
            name="gem"
            class="dashboard-validator-management-modal-search__item-gem"
          />
        </span>
        <BcTextEllipsisMiddle
          :text="`(${item.value.public_key})`"
        />
        <span class="dashboard-validator-management-modal-search__item-info">
          {{ $t('common.index') }}: {{ item.value.index }}
        </span>
      </div>
      <div
        v-if="item.type === 'validators_by_deposit_address'"
        class="dashboard-validator-management-modal-search__item"
        :class="{ 'dashboard-validator-management-modal-search__item--disabled': isDisabled(item.type) }"
      >
        <BcIcon
          name="desktop"
        />
        <span class="dashboard-validator-management-modal-search__item-validator_info">
          {{ item.value.count }} {{ $t('common.validator', item.value.count) }}
          <BcIcon
            v-if="isDisabled(item.type)"
            name="gem"
            class="dashboard-validator-management-modal-search__item-gem"
          />
        </span>
        <BcTextEllipsisMiddle
          :text="`${item.value.deposit_address}`"
        />
        <span class="dashboard-validator-management-modal-search__item-info">
          {{ $t('dashboard.validator.management.search.deposited_by') }}
        </span>
      </div>
      <div
        v-if="item.type === 'validator_list'"
        class="dashboard-validator-management-modal-search__item"
        :class="{ 'dashboard-validator-management-modal-search__item--disabled': isDisabled(item.type) }"
      >
        <BcIcon
          name="desktop"
        />
        <span class="dashboard-validator-management-modal-search__item-validator_info">
          {{ item.value.validators.length }} {{ $t('common.validator', item.value.validators.length) }}
          <BcIcon
            v-if="isDisabled(item.type)"
            name="gem"
            class="dashboard-validator-management-modal-search__item-gem"
          />
        </span>
        <BcTextEllipsis>
          {{ item.value.validators.join(', ') }}
        </BcTextEllipsis>
        <span class="dashboard-validator-management-modal-search__item-info">
          {{ $t('dashboard.validator.management.search.by_index_or_public_key') }}
        </span>
      </div>
      <div
        v-if="item.type === 'validators_by_withdrawal_credential'"
        class="dashboard-validator-management-modal-search__item"
        :class="{ 'dashboard-validator-management-modal-search__item--disabled': isDisabled(item.type) }"
      >
        <BcIcon
          name="desktop"
        />
        <span class="dashboard-validator-management-modal-search__item-validator_info">
          {{ item.value.count }} {{ $t('common.validator', item.value.count) }}
          <BcIcon
            v-if="isDisabled(item.type)"
            name="gem"
            class="dashboard-validator-management-modal-search__item-gem"
          />
        </span>
        <BcTextEllipsisMiddle
          :text="item.value.ens_name ? item.value.ens_name : item.value.withdrawal_credential"
        />
        <span class="dashboard-validator-management-modal-search__item-info">
          {{ $t('dashboard.validator.management.search.withdrawal_credential') }}
        </span>
      </div>
      <div
        v-if="item.type === 'validators_by_graffiti'"
        class="dashboard-validator-management-modal-search__item"
        :class="{ 'dashboard-validator-management-modal-search__item--disabled': isDisabled(item.type) }"
      >
        <BcIcon
          name="desktop"
        />
        <span class="dashboard-validator-management-modal-search__item-validator_info">
          {{ item.value.count }} {{ $t('common.validator', item.value.count) }}
          <BcIcon
            v-if="isDisabled(item.type)"
            name="gem"
            class="dashboard-validator-management-modal-search__item-gem"
          />
        </span>
        <BcTextEllipsis>
          {{ item.value.graffiti }}
        </BcTextEllipsis>
        <span class="dashboard-validator-management-modal-search__item-info">
          {{ $t('dashboard.validator.management.search.with_graffiti') }}
        </span>
      </div>
    </template>
  </BcInputSearch>
</template>

<style scoped lang="scss">
.dashboard-validator-management-modal-search__item-info {
  color: var(--text-color-disabled)
}
.dashboard-validator-management-modal-search__item-gem {
  color: var(--primary-color);
  isolation: isolate;
}
.dashboard-validator-management-modal-search__item {
  display: grid;
  grid-template-columns: subgrid;
  grid-column: 1/-1;
  align-items: center;
}
.dashboard-validator-management-modal-search__item--disabled {
  opacity: .7;
}
.dashboard-validator-management-modal-search__item-validator_info {
  display: flex;
  justify-content: end;
  align-items: center;
  gap: 6px;
}
</style>
