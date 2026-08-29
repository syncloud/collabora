<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api, humanSize, type FileEntry } from '../api'

const router = useRouter()
const files = ref<FileEntry[]>([])
const error = ref('')
const busy = ref(false)
const dialogOpen = ref(false)
const newName = ref('')
const newKind = ref('docx')

const kinds = [
  { kind: 'docx', label: 'Document' },
  { kind: 'xlsx', label: 'Spreadsheet' },
  { kind: 'pptx', label: 'Presentation' },
]

async function reload() {
  try {
    files.value = await api.files()
    error.value = ''
  } catch (e) {
    error.value = (e as Error).message
  }
}

function openNewDialog(kind: string) {
  newKind.value = kind
  newName.value = ''
  dialogOpen.value = true
}

async function create() {
  if (!newName.value) return
  busy.value = true
  try {
    const created = await api.create(newName.value, newKind.value)
    dialogOpen.value = false
    await router.push({ name: 'editor', params: { id: created.id } })
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    busy.value = false
  }
}

async function upload(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return
  busy.value = true
  try {
    await api.upload(file.name, file)
    await reload()
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    busy.value = false
  }
}

async function remove(entry: FileEntry) {
  busy.value = true
  try {
    await api.remove(entry.id)
    await reload()
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    busy.value = false
  }
}

function open(entry: FileEntry) {
  router.push({ name: 'editor', params: { id: entry.id } })
}

onMounted(reload)
</script>

<template>
  <div class="files" data-testid="files-view">
    <div class="actions">
      <el-dropdown trigger="click" @command="openNewDialog">
        <el-button type="primary" data-testid="new-document-button">New</el-button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item v-for="kind in kinds" :key="kind.kind" :command="kind.kind">
              <span :data-testid="`new-${kind.kind}-option`">{{ kind.label }}</span>
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>

      <label class="upload" data-testid="upload-label">
        Upload
        <input type="file" data-testid="upload-input" @change="upload" />
      </label>

      <el-button data-testid="refresh-button" @click="reload">Refresh</el-button>
    </div>

    <div v-if="error" class="error" data-testid="files-error">{{ error }}</div>

    <div v-if="files.length === 0" class="empty" data-testid="files-empty">
      No documents yet.
    </div>

    <table v-else class="list" data-testid="files-table">
      <thead>
        <tr>
          <th>Name</th>
          <th class="secondary">Kind</th>
          <th class="secondary">Size</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="entry in files" :key="entry.id" :data-testid="`file-row-${entry.name}`">
          <td class="name">
            <a href="#" :data-testid="`file-open-${entry.name}`" @click.prevent="open(entry)">
              {{ entry.name }}
            </a>
          </td>
          <td class="secondary">{{ entry.kind }}</td>
          <td class="secondary">{{ humanSize(entry.size) }}</td>
          <td class="right">
            <el-button
              size="small"
              :data-testid="`file-delete-${entry.name}`"
              @click="remove(entry)"
            >
              Delete
            </el-button>
          </td>
        </tr>
      </tbody>
    </table>

    <el-dialog v-model="dialogOpen" title="New document" class="new-dialog">
      <el-input
        v-model="newName"
        placeholder="file name"
        data-testid="new-document-name"
        @keyup.enter="create"
      />
      <template #footer>
        <el-button data-testid="new-document-cancel" @click="dialogOpen = false">Cancel</el-button>
        <el-button
          type="primary"
          :loading="busy"
          data-testid="new-document-create"
          @click="create"
        >
          Create
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.files {
  padding: 16px;
}
.actions {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
}
.upload {
  display: inline-flex;
  align-items: center;
  font-size: 14px;
  line-height: 1;
  cursor: pointer;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  padding: 0 15px;
  height: 32px;
}
.upload input {
  display: none;
}
.list {
  width: 100%;
  border-collapse: collapse;
}
.list th,
.list td {
  text-align: left;
  padding: 8px;
  border-bottom: 1px solid #ebeef5;
}
.list .right {
  text-align: right;
}
.list .name {
  word-break: break-word;
}
.empty {
  color: #909399;
  padding: 24px 8px;
}
.error {
  color: #f56c6c;
  margin-bottom: 12px;
}

@media (max-width: 640px) {
  .files {
    padding: 12px;
  }
  .actions {
    gap: 12px;
  }
  .secondary {
    display: none;
  }
  .list th,
  .list td {
    padding: 12px 8px;
  }
}
</style>
