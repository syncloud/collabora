<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api } from './api'

const username = ref('')

onMounted(async () => {
  const session = await api.session()
  username.value = session.name || session.username
})
</script>

<template>
  <div class="app" data-testid="app">
    <header class="bar">
      <span class="title" data-testid="app-title">Collabora Online</span>
      <span class="spacer" />
      <span v-if="username" class="user" data-testid="current-user">{{ username }}</span>
      <a class="logout" href="/oidc/logout" data-testid="logout-link">Sign out</a>
    </header>
    <main>
      <router-view />
    </main>
  </div>
</template>

<style>
body {
  margin: 0;
  font-family: system-ui, sans-serif;
}
.app {
  display: flex;
  flex-direction: column;
  height: 100vh;
  height: 100dvh;
}
.bar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 16px;
  border-bottom: 1px solid #dcdfe6;
}
.title {
  font-weight: 600;
}
.spacer {
  flex: 1;
}
.user {
  color: #606266;
  font-size: 14px;
}
.logout {
  color: #409eff;
  font-size: 14px;
  text-decoration: none;
}
main {
  flex: 1;
  min-height: 0;
}
</style>
