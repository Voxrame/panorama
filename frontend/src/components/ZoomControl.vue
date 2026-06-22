<script setup lang="ts">
defineEmits<{
  zoomIn: []
  zoomOut: []
}>()
</script>

<template>
  <div class="flex flex-col bg-white shadow-lg divide-y rounded-lg">
    <button class="relative w-8 h-8 font-semibold" type="button" @click="$emit('zoomIn')">&plus;</button>
    <button class="relative w-8 h-8 font-semibold" type="button" @click="$emit('zoomOut')">&minus;</button>
  </div>
</template>
