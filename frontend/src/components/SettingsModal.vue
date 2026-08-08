<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto bg-black/40 p-4 sm:items-center"
      @click.self="$emit('close')"
    >
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="settings-title"
        class="my-auto w-full max-w-lg rounded-xl border border-border bg-card shadow-xl"
      >
        <!-- header -->
        <div class="flex items-center justify-between border-b border-border px-4 py-3">
          <h2 id="settings-title" class="text-sm font-semibold text-ink">Speedtest Server</h2>
          <button
            type="button"
            class="inline-flex h-8 w-8 items-center justify-center rounded-lg text-muted transition-colors hover:bg-surface hover:text-ink"
            aria-label="Close"
            @click="$emit('close')"
          >
            <svg class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
              <path d="M18 6 6 18M6 6l12 12" />
            </svg>
          </button>
        </div>

        <!-- body -->
        <div class="max-h-[60vh] overflow-y-auto px-4 py-3">
          <div class="flex flex-col gap-1">
            <!-- Automatic -->
            <label
              class="flex cursor-pointer items-center gap-3 rounded-lg border border-transparent px-3 py-2 hover:bg-surface"
              :class="pick === 'auto' ? 'border-accent bg-accent/10' : ''"
            >
              <input type="radio" name="server" class="accent-accent" :value="'auto'" v-model="pick" />
              <div>
                <div class="text-sm font-medium text-ink">Automatic (closest)</div>
                <div class="text-xs text-muted">Let the Ookla CLI pick the nearest server each run.</div>
              </div>
            </label>

            <!-- Loading / error for the server list -->
            <div v-if="loadingServers" class="flex items-center gap-2 px-3 py-2 text-sm text-muted">
              <Spinner /> Loading nearby servers…
            </div>
            <div v-else-if="serverError" class="px-3 py-2 text-sm text-danger">
              {{ serverError }}
            </div>

            <!-- Nearest servers -->
            <label
              v-for="s in servers"
              :key="s.id"
              class="flex cursor-pointer items-center gap-3 rounded-lg border border-transparent px-3 py-2 hover:bg-surface"
              :class="pick === s.id ? 'border-accent bg-accent/10' : ''"
            >
              <input type="radio" name="server" class="accent-accent" :value="s.id" v-model="pick" />
              <div class="min-w-0">
                <div class="truncate text-sm font-medium text-ink">{{ s.name }}</div>
                <div class="truncate text-xs text-muted">
                  {{ s.location }}<template v-if="s.country">, {{ s.country }}</template>
                  · #{{ s.id }}<template v-if="s.host"> · {{ s.host }}</template>
                </div>
              </div>
            </label>

            <!-- Manual server ID -->
            <label
              class="flex cursor-pointer items-center gap-3 rounded-lg border border-transparent px-3 py-2 hover:bg-surface"
              :class="pick === 'manual' ? 'border-accent bg-accent/10' : ''"
            >
              <input type="radio" name="server" class="accent-accent" :value="'manual'" v-model="pick" />
              <div class="flex flex-1 flex-wrap items-center gap-2">
                <div>
                  <div class="text-sm font-medium text-ink">Manual server ID</div>
                  <div class="text-xs text-muted">Pin any Ookla server by numeric ID.</div>
                </div>
                <input
                  v-model="manualId"
                  type="number"
                  min="1"
                  inputmode="numeric"
                  placeholder="e.g. 12345"
                  class="ml-auto w-28 rounded-lg border border-border bg-surface px-2 py-1 text-sm text-ink tabular outline-none focus:border-accent"
                  @focus="pick = 'manual'"
                />
              </div>
            </label>
          </div>
        </div>

        <!-- footer -->
        <div class="flex items-center justify-end gap-2 border-t border-border px-4 py-3">
          <p v-if="saveError" class="mr-auto text-sm text-danger">{{ saveError }}</p>
          <button
            type="button"
            class="rounded-lg border border-border px-3 py-2 text-sm font-medium text-muted transition-colors hover:bg-surface hover:text-ink"
            @click="$emit('close')"
          >
            Cancel
          </button>
          <button
            type="button"
            class="inline-flex items-center gap-2 rounded-lg bg-accent px-3 py-2 text-sm font-medium text-white transition-colors hover:bg-accent-strong disabled:opacity-60"
            :disabled="saving || !canSave"
            @click="save"
          >
            <Spinner v-if="saving" />
            Save
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import Spinner from '@/components/Spinner.vue'
import { getServers, getSettings, updateSettings } from '@/services/api'
import type { Server } from '@/types/result'

const props = defineProps<{ open: boolean }>()
const emit = defineEmits<{ close: [] }>()

const servers = ref<Server[]>([])
const loadingServers = ref(false)
const serverError = ref('')

// 'auto' | 'manual' | <server id>
const pick = ref<'auto' | 'manual' | number>('auto')
const manualId = ref('')

const saving = ref(false)
const saveError = ref('')

const canSave = computed(() => {
  if (pick.value === 'manual') return Number(manualId.value) > 0
  return true
})

watch(
  () => props.open,
  (open) => {
    if (open) load()
  },
)

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && props.open) emit('close')
}
onMounted(() => window.addEventListener('keydown', onKeydown))
onUnmounted(() => window.removeEventListener('keydown', onKeydown))

async function load() {
  saveError.value = ''
  serverError.value = ''
  loadingServers.value = true

  // Settings is a fast DB read; servers shells out to the CLI (~seconds).
  // Fetch both concurrently, keeping their independent failure handling.
  const [settingsRes, serversRes] = await Promise.allSettled([getSettings(), getServers()])
  loadingServers.value = false

  const selectedId = settingsRes.status === 'fulfilled' ? settingsRes.value.server_id : null

  if (serversRes.status === 'fulfilled') {
    servers.value = serversRes.value
  } else {
    serverError.value = 'Could not load nearby servers. You can still choose Automatic or a manual ID.'
    servers.value = []
  }

  manualId.value = ''
  if (selectedId == null) {
    pick.value = 'auto'
  } else if (servers.value.some((s) => s.id === selectedId)) {
    pick.value = selectedId
  } else {
    // A pinned id that isn't in the nearest list — surface it as manual.
    pick.value = 'manual'
    manualId.value = String(selectedId)
  }
}

async function save() {
  saving.value = true
  saveError.value = ''
  try {
    if (pick.value === 'auto') {
      await updateSettings({ server_id: null })
    } else if (pick.value === 'manual') {
      await updateSettings({ server_id: Number(manualId.value) })
    } else {
      const s = servers.value.find((x) => x.id === pick.value)
      await updateSettings({
        server_id: pick.value,
        server_name: s?.name,
        server_location: s?.location,
      })
    }
    emit('close')
  } catch (err) {
    saveError.value = `Failed to save: ${(err as Error).message}`
  } finally {
    saving.value = false
  }
}
</script>
