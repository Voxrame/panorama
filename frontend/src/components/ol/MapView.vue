<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import 'ol/ol.css'
import Map from 'ol/Map'
import View from 'ol/View'
import TileLayer from 'ol/layer/Tile'
import XYZ from 'ol/source/XYZ'
import Projection from 'ol/proj/Projection'
import TileGrid from 'ol/tilegrid/TileGrid'
import type MapBrowserEvent from 'ol/MapBrowserEvent'
import { worldPositionUnderCursor } from '@/stores/worldPosition'

const mapElement = ref<HTMLDivElement>()

const ORIGIN_WORLD: [number, number] = [735, -818]

function projectWorld(x: number, z: number): [number, number] {
  return [((z - x) * 16) / 2, ((z + x) * 16) / 4]
}

const [originIsoX, originIsoY] = projectWorld(...ORIGIN_WORLD)

const ZOOM_OFFSET = 8
const RESOLUTIONS = Array.from({ length: ZOOM_OFFSET + 1 }, (_, i) => 2 ** (ZOOM_OFFSET - i))
const MIN_LEVEL = 0
const MAX_LEVEL = RESOLUTIONS.length - 1

const clamp = (n: number, min: number, max: number) => Math.max(min, Math.min(max, n))

const projection = new Projection({
  code: 'panorama-iso',
  units: 'pixels',
  extent: [-65536, -32768, 65536, 32768],
})

const tileGrid = new TileGrid({
  origin: [0, 0],
  resolutions: RESOLUTIONS,
  tileSize: 256,
})

let map: Map | null = null

function updateCoordinates(event: MapBrowserEvent<UIEvent>) {
  const cx = event.coordinate[0]
  const cy = event.coordinate[1]
  if (cx === undefined || cy === undefined) return
  // Inverse isometric projection (coordY is y-up, hence -2*cy).
  const x = (-2 * cy - cx) / 16
  const z = (-2 * cy + cx) / 16
  worldPositionUnderCursor.value = { x: Math.round(x), z: Math.round(z) }
}

function resolutionToLevel(resolution: number): number {
  return clamp(Math.round(ZOOM_OFFSET - Math.log2(resolution)), MIN_LEVEL, MAX_LEVEL)
}

function stepZoom(deltaLevels: number) {
  const view = map?.getView()
  if (!view) return
  const resolution = view.getResolution()
  if (resolution === undefined) return
  const target = clamp(resolutionToLevel(resolution) + deltaLevels, MIN_LEVEL, MAX_LEVEL)
  view.animate({ resolution: RESOLUTIONS[target], duration: 250 })
}

function zoomIn() {
  stepZoom(1)
}

function zoomOut() {
  stepZoom(-1)
}

defineExpose({ zoomIn, zoomOut })

onMounted(() => {
  if (!mapElement.value) return

  map = new Map({
    target: mapElement.value,
    controls: [],
    layers: [
      new TileLayer({
        source: new XYZ({
          projection,
          tileGrid,
          tileUrlFunction: (tileCoord) => {
            const [olZ, x, y] = tileCoord
            if (olZ === undefined || x === undefined || y === undefined) return undefined
            const serverZ = olZ - ZOOM_OFFSET
            return `/tiles/${serverZ}/${x}/${y}.png`
          },
        }),
      }),
    ],
    view: new View({
      projection,
      center: [originIsoX, -originIsoY],
      resolution: RESOLUTIONS[MAX_LEVEL],
      resolutions: RESOLUTIONS,
    }),
  })

  map.on('pointermove', updateCoordinates)
  map.on('click', updateCoordinates)
})

onBeforeUnmount(() => {
  map?.setTarget(undefined)
  map = null
})
</script>

<template>
  <div ref="mapElement" class="w-full h-full bg-black" />
</template>
