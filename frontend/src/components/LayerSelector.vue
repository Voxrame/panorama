<script setup lang="ts">
import { ref } from 'vue'

const modes = [
  { id: 'isometric', name: '3D' },
  { id: 'topdown', name: '2D' },
] as const

const selected = ref<(typeof modes)[number]['id']>('isometric')
</script>

<template>
  <div class="flex bg-white shadow-lg divide-x rounded-lg">
    <button
      v-for="(mode, id) in modes"
      :key="mode.id"
      type="button"
      class="px-2 py-1 font-semibold"
      :class="[
        id === 0 ? 'rounded-l-lg' : '',
        id === modes.length - 1 ? 'rounded-r-lg' : '',
        selected === mode.id ? 'bg-red-600 text-white' : '',
      ]"
      @click="selected = mode.id"
    >
      {{ mode.name }}
    </button>
  </div>
</template>
