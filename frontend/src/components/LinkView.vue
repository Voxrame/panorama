<script setup lang="ts">
import LinkIcon from './icons/LinkIcon.vue'

function linkView() {
  alert()
}
</script>

<template>
  <button class="flex px-1 py-1 bg-white border shadow-lg rounded-lg" type="button" @click="linkView">
    <LinkIcon class="w-6 h-6" />
  </button>
</template>
