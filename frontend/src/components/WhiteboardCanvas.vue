<template>
  <div class="wrapper">
    <canvas ref="canvasEl" class="board"></canvas>
    <div class="placeholder">白板初始化中 (Fabric)</div>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue';
import { fabric } from 'fabric';

const canvasEl = ref<HTMLCanvasElement | null>(null);
let canvas: fabric.Canvas | null = null;

onMounted(() => {
  if (!canvasEl.value) return;
  canvas = new fabric.Canvas(canvasEl.value, {
    backgroundColor: '#f8fafc',
    selection: true
  });
  canvas.setWidth(canvasEl.value.clientWidth);
  canvas.setHeight(500);
});

onBeforeUnmount(() => {
  canvas?.dispose();
  canvas = null;
});
</script>

<style scoped>
.wrapper {
  position: relative;
  width: 100%;
  height: 100%;
  min-height: 500px;
}

.board {
  width: 100%;
  height: 100%;
  display: block;
}

.placeholder {
  position: absolute;
  top: 8px;
  right: 12px;
  padding: 4px 8px;
  background: rgba(15, 23, 42, 0.72);
  color: #e2e8f0;
  border-radius: 6px;
  font-size: 12px;
}
</style>
