<template>
  <div class="chat">
    <header>聊天室</header>
    <section class="messages">
      <div class="msg system">等待接入 WebSocket…</div>
    </section>
    <footer>
      <input type="text" placeholder="输入消息..." />
      <button>发送</button>
    </footer>
  </div>
</template>

<script setup lang="ts"></script>

<style scoped>
.chat {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 500px;
}

header {
  padding: 12px;
  font-weight: 700;
  border-bottom: 1px solid #e2e8f0;
}

.messages {
  flex: 1;
  padding: 12px;
  overflow-y: auto;
  background: #f8fafc;
}

.msg.system {
  color: #475569;
  font-size: 13px;
}

footer {
  display: flex;
  gap: 8px;
  padding: 12px;
  border-top: 1px solid #e2e8f0;
  background: #fff;
}

input {
  flex: 1;
  padding: 10px 12px;
  border-radius: 8px;
  border: 1px solid #cbd5e1;
  outline: none;
}

button {
  padding: 10px 14px;
  border: none;
  border-radius: 8px;
  background: #2563eb;
  color: #fff;
  cursor: pointer;
  font-weight: 600;
}
</style>
