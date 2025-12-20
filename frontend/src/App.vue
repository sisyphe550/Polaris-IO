<template>
  <div class="page">
    <header class="header">
      <div class="brand">Shared Board</div>
      <div class="room-actions">
        <button>创建房间</button>
        <button class="ghost">加入房间</button>
      </div>
    </header>
    <main class="layout">
      <section class="canvas-area">
        <WhiteboardCanvas />
      </section>
      <aside class="chat-area">
        <ChatPanel />
      </aside>
    </main>
  </div>
</template>

<script setup lang="ts">
import WhiteboardCanvas from './components/WhiteboardCanvas.vue';
import ChatPanel from './components/ChatPanel.vue';
</script>

<style scoped>
.page {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: #0f172a;
  color: #e2e8f0;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
}

.brand {
  font-weight: 700;
  letter-spacing: 0.4px;
}

.room-actions button {
  margin-left: 8px;
  padding: 8px 12px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 600;
}

.room-actions .ghost {
  background: transparent;
  color: #e2e8f0;
  border: 1px solid #334155;
}

.layout {
  flex: 1;
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: 12px;
  padding: 12px;
}

.canvas-area,
.chat-area {
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(15, 23, 42, 0.08);
  overflow: hidden;
}

.canvas-area {
  min-height: 520px;
}
</style>
