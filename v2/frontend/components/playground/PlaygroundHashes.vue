<script setup lang="ts">
const data = {
  data: [
    {
      ens: 'keinMini.ens',
      hash: '0xE276Bc378A527A8792B353cdCA5b5E53263DfB9e',
      type: 'address',
    },
    {
      ens: 'thisIsAVerryLongEnsNameSoItShoulBeTruncatedthisIsAVerryLongEnsNameSoItShoulBeTruncated.ens',
      hash: '0xE276Bc378A527A8792B353cdCA5b5E53263DfB9e',
      type: 'address',
    },
    {
      hash: '0xE276Bc378A527A8792B353cdCA5b5E53263DfB9e',
      type: 'address',
    },
    {
      hash: '0x010000000000000000000000e276bc378a527a8792b353cdca5b5e53263dfb9e',
      type: 'withdrawal_credentials',
    },
    {
      hash: '0x0094ddaa84a9dce074b5abaa0e939402029a4d3f5f9eea57986583630353ea49',
      type: 'withdrawal_credentials',
    },
    {
      hash: '0x9340bfc34ffab8c28b1870a4125c559978ac2b278f76f462b5c859a00c3ba3426b176dc2c689096ad575b4cd4dbb76ae',
      type: 'public_key',
    },
    {
      hash: '0xbc4275221772ede9869d9cf17bf4b5de5a8fe73aa63657abfa9c8b62f3ef39af',
      type: 'block_hash',
    },
    {
      hash: '0x4d9e52c17e163fe6e542dcf3ebdcc53b74f5a191c017bc153a9857aaa66683ee',
      type: 'root',
    },
    { hash: '0x910fc508a2d992748628d9fa800a4055828fa3bb200069dc00f0f29eb1d607e20dd645c4339f48a5dd468fb51f15c9d20c5a161b1219fa9ed2d48aa8140eca56ceafd691df2d1037411ff2fb8b9912ca215ebff15d41a40eebb496b76fe403f4' },
  ].map((item, index) => ({
    ...item, index,
  })),
  paging: { total_count: 1000 },
}
</script>

<template>
  <BcTable
    class="hash-table"
    :data
    data-key="index"
    :page-size="50"
  >
    <Column
      field="type"
      header="Type"
    />
    <Column
      field="hash"
      header="Hash"
    >
      <template #body="{ data: { hash } }">
        <div class="hash">
          {{ hash }}
        </div>
      </template>
    </Column>
    <Column
      field="short"
      header="Short"
    >
      <template #body="{ data: { hash, ens, type } }">
        <BcFormatHash
          :hash
          :type
          :ens
        />
      </template>
    </Column>
    <Column
      field="full"
      header="Full"
    >
      <template #body="{ data: { hash, ens, type } }">
        <BcFormatHash
          :hash
          :type
          :ens
          :full="true"
        />
      </template>
    </Column>
    <Column
      field="short_no_link"
      header="Short no link"
    >
      <template #body="{ data: { hash, ens, type } }">
        <BcFormatHash
          :hash
          :type
          :ens
          :no-link="true"
        />
      </template>
    </Column>
    <Column
      field="full_no_link"
      header="Full no link"
    >
      <template #body="{ data: { hash, ens, type } }">
        <BcFormatHash
          :hash
          :type
          :ens
          :full="true"
          :no-link="true"
        />
      </template>
    </Column>
    <Column
      field="short_no_copy"
      header="Short no copy"
    >
      <template #body="{ data: { hash, ens, type } }">
        <BcFormatHash
          :hash
          :type
          :ens
          :no-copy="true"
        />
      </template>
    </Column>
    <Column
      field="full_no_copy"
      header="Full no copy"
    >
      <template #body="{ data: { hash, ens, type } }">
        <BcFormatHash
          :hash
          :type
          :ens
          :full="true"
          :no-copy="true"
        />
      </template>
    </Column>
  </BcTable>
</template>

<style lang="scss" scoped>
.hash {
  word-wrap: break-word;
}

.hash-table {
  :deep(td) {
    max-width: 200px;
    width: 200px;
  }
}
</style>
