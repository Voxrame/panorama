<script setup lang="ts">
import type { WorldCursorPosition } from '@/stores/worldPosition'

defineProps<{
  position: WorldCursorPosition | null
}>()
</script>

<template>
  <div class="flex bg-white border shadow-lg rounded-lg">
    <template v-if="position">
      <div class="px-2 py-1">
        X: <span class="font-semibold font-mono">{{ position.x }}</span>
      </div>
      <div class="pr-2 py-1">
        Z: <span class="font-semibold font-mono">{{ position.z }}</span>
      </div>
    </template>
    <div v-else class="px-2 py-1">Coordinates will be shown here</div>
  </div>
</template>
