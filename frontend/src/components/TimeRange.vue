<template>
  <div role="tablist" class="tabs tabs-box">
    <button
      v-for="opt in options"
      :key="opt.key"
      role="tab"
      class="tab"
      :class="{ 'tab-active': opt.key === modelValue }"
      @click="$emit('update:modelValue', opt.key)"
    >
      {{ opt.label }}
    </button>
  </div>
</template>

<script setup lang="ts">
export type RangeKey = '24h' | '7d' | '30d' | 'all'

defineProps<{ modelValue: RangeKey }>()
defineEmits<{ 'update:modelValue': [value: RangeKey] }>()

const options: { key: RangeKey; label: string }[] = [
  { key: '24h', label: '24h' },
  { key: '7d', label: '7d' },
  { key: '30d', label: '30d' },
  { key: 'all', label: 'All' },
]
</script>
