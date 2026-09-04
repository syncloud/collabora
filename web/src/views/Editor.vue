<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../api'

const props = defineProps<{ id: string }>()
const router = useRouter()

const frameName = 'collabora-editor'
const loading = ref(true)
const error = ref('')
const title = ref('')

function submitToFrame(url: string, token: string, ttl: number) {
  const form = document.createElement('form')
  form.action = url
  form.method = 'POST'
  form.target = frameName
  form.style.display = 'none'

  for (const [name, value] of [
    ['access_token', token],
    ['access_token_ttl', String(ttl)],
  ]) {
    const input = document.createElement('input')
    input.type = 'hidden'
    input.name = name
    input.value = value
    form.appendChild(input)
  }

  document.body.appendChild(form)
  form.submit()
  document.body.removeChild(form)
}

function back() {
  router.push({ name: 'files' })
}

onMounted(async () => {
  try {
    const config = await api.editor(props.id)
    title.value = config.name
    submitToFrame(config.url, config.token, config.ttl)
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="editor" data-testid="editor-view">
    <div class="bar">
      <el-button size="small" data-testid="editor-back" @click="back">Back</el-button>
      <span class="name" data-testid="editor-title">{{ title }}</span>
    </div>

    <div v-if="loading" class="notice" data-testid="editor-loading">Opening document...</div>
    <div v-if="error" class="notice error" data-testid="editor-error">{{ error }}</div>

    <iframe
      :name="frameName"
      class="frame"
      data-testid="editor-frame"
      allow="clipboard-read; clipboard-write"
    />
  </div>
</template>

<style scoped>
.editor {
  display: flex;
  flex-direction: column;
  height: 100%;
}
.bar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 16px;
  border-bottom: 1px solid #ebeef5;
}
.name {
  font-size: 14px;
  color: #606266;
}
.frame {
  flex: 1;
  width: 100%;
  border: 0;
  min-height: 0;
}
.notice {
  padding: 8px 16px;
  color: #909399;
  font-size: 14px;
}
.notice.error {
  color: #f56c6c;
}
</style>
