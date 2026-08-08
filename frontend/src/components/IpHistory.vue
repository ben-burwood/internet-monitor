<template>
  <div class="rounded-xl border border-border bg-card">
    <div class="flex flex-wrap items-center justify-between gap-2 border-b border-border px-4 py-3">
      <h2 class="text-sm font-semibold text-ink">Public IP history</h2>
      <span class="text-xs text-muted">{{ changeLabel }}</span>
    </div>
    <div class="px-4 py-3">
      <p v-if="segments.length === 0" class="text-sm text-muted">No public IP recorded yet.</p>
      <ol v-else class="flex flex-col">
        <li
          v-for="(seg, i) in segments"
          :key="seg.first_seen + seg.ip"
          class="relative flex gap-3 pb-4 last:pb-0"
        >
          <!-- timeline rail -->
          <span
            v-if="i !== segments.length - 1"
            class="absolute left-[3.5px] top-3 h-full w-px bg-border"
            aria-hidden="true"
          ></span>
          <span
            class="relative mt-1.5 h-2 w-2 shrink-0 rounded-full"
            :class="i === 0 ? 'bg-accent' : 'bg-muted'"
          ></span>

          <div class="min-w-0">
            <div class="flex flex-wrap items-center gap-2">
              <span class="font-semibold tabular text-ink">{{ seg.ip }}</span>
              <span
                v-if="i === 0"
                class="rounded bg-accent/15 px-1.5 py-0.5 text-xs font-medium text-accent"
              >
                current
              </span>
              <span v-if="seg.isp" class="text-xs text-muted">{{ seg.isp }}</span>
            </div>
            <div class="mt-0.5 text-xs text-muted">
              {{ formatDate(seg.first_seen) }} – {{ i === 0 ? 'now' : formatDate(seg.last_seen) }}
              · {{ seg.count.toLocaleString() }} test{{ seg.count === 1 ? '' : 's' }}
            </div>
          </div>
        </li>
      </ol>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { IPSegment } from '@/types/result'
import { formatDate } from '@/utils/format'

const props = defineProps<{ segments: IPSegment[] }>()

const changeLabel = computed(() => {
  const changes = Math.max(0, props.segments.length - 1)
  if (props.segments.length === 0) return ''
  return changes === 0 ? 'No changes' : `${changes} change${changes === 1 ? '' : 's'}`
})

</script>
