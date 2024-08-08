<script setup lang="ts">
interface Props {
  ens?: string
  full?: boolean // if true the hash will not be truncated
  hash?: string
  noCopy?: boolean // same as for the link
  // most of the time we want to render it as a link (if possible), but there might be cases where we don't
  noLink?: boolean
  noWrap?: boolean // don't wrap elements
  type?:
    | 'address'
    | 'block_hash'
    | 'public_key'
    | 'root' // if none is provided the default format will be applied
    | 'tx'
    | 'withdrawal_credentials'
}
const props = defineProps<Props>()

const data = computed(() => {
  if (!props.hash || props.hash === '0x') {
    return
  }
  const hash = props.hash
  const className = props.full
    ? 'full'
    : props.ens
      ? 'truncate-text'
      : props.noWrap
        ? 'no-wrap'
        : ''
  let parts: { className?: string, value: string }[] = []
  let link: string = ''
  if (props.ens) {
    parts.push({
      className: !props.full ? 'truncate-text' : '',
      value: props.ens,
    })
  }
  else if (props.type === 'withdrawal_credentials') {
    const isSet = hash.startsWith('0x01')
    const color = isSet ? 'green' : 'orange'
    parts.push({ className: color, value: hash.substring(0, 4) })
    if (props.full) {
      parts.push({ value: hash.substring(4) })
    }
    else {
      parts = parts.concat([
        { className: 'dots-before', value: hash.substring(26, 30) },
        { className: 'dots-before', value: hash.substring(hash.length - 4) },
      ])
    }
    if (isSet && !props.noLink) {
      link = `/address/0x${props.hash.substring(26)}`
    }
  }
  else {
    const color = props.full ? 'prime' : undefined
    const middle = props.full
      ? { value: hash.substring(6, hash.length - 4) }
      : { className: 'dots-before', value: '' }
    parts = [
      { value: '0x' },
      { className: color, value: hash.substring(2, 6) },
      middle,
      { className: color, value: hash.substring(hash.length - 4) },
    ]
  }
  if (!props.noLink) {
    switch (props.type) {
      case 'address':
        link = `/address/${props.hash}`
        break
      case 'block_hash':
        link = `/block/${props.hash}`
        break
      case 'root':
        link = `/slot/${props.hash}`
        break
      case 'public_key':
        link = `/validator/${props.hash}`
        break
      case 'tx':
        link = `/tx/${props.hash}`
        break
    }
  }

  return {
    className,
    link,
    parts,
  }
})
</script>

<template>
  <div
    v-if="data"
    class="format-hash"
    :class="{ 'no-wrap': noWrap }"
  >
    <BcTooltip class="tt-container">
      <template
        v-if="!full || ens"
        #tooltip
      >
        <div
          v-if="ens"
          class="tt ens-name full"
        >
          {{ ens }}
        </div>
        <div class="tt">
          <BcFormatHash
            :hash="hash"
            :full="true"
            :no-link="true"
            :no-copy="true"
            :type="type"
          />
        </div>
      </template>
      <span
        v-if="!data.link"
        :class="data.className"
      >
        <span
          v-for="(part, index) in data.parts"
          :key="index"
          :class="part.className"
        >
          {{ part.value }}
        </span>
      </span>
      <BcLink
        v-else
        :to="data.link"
        target="_blank"
        class="link"
        :class="data.className"
      >
        <span
          v-for="(part, index) in data.parts"
          :key="index"
          :class="part.className"
        >
          {{ part.value }}
        </span>
      </BcLink>
    </BcTooltip>
    <BcCopyToClipboard
      v-if="!props.noCopy"
      :value="props.hash"
      class="copy"
    />
  </div>
</template>

<style lang="scss" scoped>
.no-wrap {
  display: flex;
  flex-wrap: nowrap;
}

.format-hash {
  &:has(.truncate-text) {
    display: flex;
  }

  &:has(.no-wrap) {
    display: flex;
    flex-wrap: nowrap;
  }

  .tt-container {
    &:has(.truncate-text) {
      display: flex;
      overflow: hidden;
    }

    &:has(.no-wrap) {
      display: flex;
      flex-wrap: nowrap;
    }
  }

  .prime {
    color: var(--primary-color);
  }

  .green {
    color: var(--positive-color);
  }

  .orange {
    color: var(--orange-color);
  }

  .copy {
    margin-left: var(--padding);
    line-height: 100%;
  }
}

.full {
  word-wrap: break-word;
}

.tt {
  max-width: 300px;
  text-align: left;

  :deep(.format-hash) {
    .green {
      color: var(--light-green);
    }

    .orange {
      color: var(--light-orange);
    }
  }
}

.ens-name {
  margin-bottom: var(--padding-small);
}
</style>
