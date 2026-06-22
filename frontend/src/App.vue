<script setup lang="ts">
import { ref } from 'vue'
import MapView from './components/ol/MapView.vue'
import ZoomControl from './components/ZoomControl.vue'
import CoordinatesView from './components/CoordinatesView.vue'
import { worldPositionUnderCursor } from './stores/worldPosition'

const mapView = ref<InstanceType<typeof MapView>>()
</script>

<template>
  <div class="relative h-screen">
    <MapView ref="mapView" />

    <div class="absolute top-4 right-4 z-[1000]">
      <ZoomControl @zoom-in="mapView?.zoomIn()" @zoom-out="mapView?.zoomOut()" />
    </div>

    <div class="absolute top-4 left-4 z-[1000]">
      <CoordinatesView :position="worldPositionUnderCursor" />
    </div>
  </div>
</template>
