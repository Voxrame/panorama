import { ref } from 'vue'

export interface WorldCursorPosition {
  x: number
  z: number
}

export const worldPositionUnderCursor = ref<WorldCursorPosition | null>(null)
