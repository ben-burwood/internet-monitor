<template>
  <div role="tablist" class="inline-flex rounded-lg border border-border bg-card p-0.5">
    <button
      v-for="opt in options"
      :key="opt.key"
      type="button"
      role="tab"
      :aria-selected="opt.key === modelValue"
      class="rounded-md px-3 py-1 text-sm font-medium transition-colors"
      :class="
        opt.key === modelValue
          ? 'bg-accent text-white'
          : 'text-muted hover:text-ink'
      "
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
