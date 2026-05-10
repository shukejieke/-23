<script setup>
import { computed, onMounted, ref } from 'vue'
import { ApiGetBusinessAvailability, ApiGetRawCaptureList } from '../api'

const loading = ref(false)
const items = ref([])
const summary = ref({})
const captures = ref([])
const cmdId = ref('103')

const readyCount = computed(() => items.value.filter((x) => x.ready).length)

async function loadAll() {
  loading.value = true
  try {
    const [availabilityRes, captureRes] = await Promise.all([
      ApiGetBusinessAvailability(),
      ApiGetRawCaptureList({ cmd_id: Number(cmdId.value), limit: 30 }),
    ])
    items.value = availabilityRes?.data?.data?.items || []
    summary.value = availabilityRes?.data?.data?.summary || {}
    captures.value = captureRes?.data?.data?.items || []
  } finally {
    loading.value = false
  }
}

function fmtTime(ts) {
  if (!ts) return '-'
  const d = new Date(Number(ts) * 1000)
  if (Number.isNaN(d.getTime())) return String(ts)
  return d.toLocaleString('zh-CN', { hour12: false })
}

onMounted(loadAll)
</script>

<template>
  <div class="page">
    <div class="header">
      <div>
        <div class="title">业务数据可得性面板</div>
        <div class="subtitle">先确认 `103/92` 是否稳定产出，再谈功能复现</div>
      </div>
      <div class="actions">
        <a class="button" @click="loadAll">刷新</a>
      </div>
    </div>

    <div class="summary">
      <div class="summary-card">
        <div class="summary-label">已就绪模块</div>
        <div class="summary-value">{{ readyCount }}/{{ items.length }}</div>
      </div>
      <div class="summary-card">
        <div class="summary-label">成员表</div>
        <div class="summary-value">{{ summary.team_user_count || 0 }}</div>
      </div>
      <div class="summary-card">
        <div class="summary-label">普通战报</div>
        <div class="summary-value">{{ summary.report_count || 0 }}</div>
      </div>
      <div class="summary-card">
        <div class="summary-label">详细战报</div>
        <div class="summary-value">{{ summary.battle_count || 0 }}</div>
      </div>
      <div class="summary-card">
        <div class="summary-label">原始 103/92</div>
        <div class="summary-value">{{ summary.raw_103_count || 0 }}/{{ summary.raw_92_count || 0 }}</div>
      </div>
    </div>

    <div class="panel">
      <div class="panel-title">模块状态</div>
      <table class="table">
        <thead>
          <tr>
            <th>模块</th>
            <th>状态</th>
            <th>原因</th>
            <th>计数</th>
            <th>依赖</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in items" :key="item.key">
            <td>{{ item.title }}</td>
            <td>
              <span :class="['status', item.status === 'ready' ? 'ok' : 'bad']">
                {{ item.status === 'ready' ? '已就绪' : '阻塞' }}
              </span>
            </td>
            <td>{{ item.reason }}</td>
            <td><pre class="mini">{{ JSON.stringify(item.counts, null, 2) }}</pre></td>
            <td>{{ (item.dependencies || []).join(' | ') }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="panel">
      <div class="panel-head">
        <div class="panel-title">103/92 原始包归档</div>
        <div class="filter">
          <select v-model="cmdId" @change="loadAll">
            <option value="103">103 成员主协议</option>
            <option value="92">92 战报主协议</option>
          </select>
        </div>
      </div>
      <div v-if="captures.length === 0" class="empty">
        当前没有归档记录，说明还没稳定抓到该主协议。
      </div>
      <div v-for="cap in captures" :key="cap.id" class="capture-card">
        <div class="capture-head">
          <div>#{{ cap.id }} cmd={{ cap.cmd_id }} {{ cap.cmd_name }}</div>
          <div>{{ fmtTime(cap.capture_time) }}</div>
        </div>
        <div class="capture-meta">
          <span>data_type={{ cap.data_type }}</span>
          <span>payload={{ cap.payload_len }}</span>
          <span>buf={{ cap.packet_buf_len }}</span>
          <span>{{ cap.src }} -> {{ cap.dst }}</span>
        </div>
        <div class="code-title">preview</div>
        <pre class="code">{{ cap.preview || '-' }}</pre>
        <div class="code-title">decoded</div>
        <pre class="code">{{ cap.decoded_text || '-' }}</pre>
      </div>
    </div>
  </div>
</template>

<style scoped>
.page {
  min-height: 100vh;
  padding: 20px;
  background: linear-gradient(180deg, #f7f8fb 0%, #eef2f9 100%);
}

.header,
.panel-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.title {
  font-size: 24px;
  font-weight: 800;
  color: #1f2f4a;
}

.subtitle {
  margin-top: 6px;
  color: #5a6780;
}

.actions {
  display: flex;
  gap: 10px;
}

.button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 36px;
  padding: 0 16px;
  border-radius: 8px;
  background: #1b5ed7;
  color: #fff;
  text-decoration: none;
  cursor: pointer;
}

.button.ghost {
  background: #fff;
  color: #274062;
  border: 1px solid #c9d5e6;
}

.summary {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 12px;
  margin-top: 16px;
}

.summary-card,
.panel,
.capture-card {
  background: rgba(255,255,255,0.94);
  border: 1px solid #d6dfed;
  border-radius: 12px;
  box-shadow: 0 10px 30px rgba(21, 42, 77, 0.06);
}

.summary-card {
  padding: 14px;
}

.summary-label {
  color: #6a7891;
  font-size: 13px;
}

.summary-value {
  margin-top: 8px;
  color: #1f2f4a;
  font-size: 28px;
  font-weight: 800;
}

.panel {
  margin-top: 16px;
  padding: 16px;
}

.panel-title {
  font-size: 18px;
  font-weight: 700;
  color: #243754;
}

.table {
  width: 100%;
  margin-top: 12px;
  border-collapse: collapse;
}

.table th,
.table td {
  padding: 12px;
  border: 1px solid #d6dfed;
  text-align: left;
  vertical-align: top;
}

.table th {
  background: #f2f6fc;
}

.status {
  display: inline-flex;
  padding: 4px 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 700;
}

.status.ok {
  background: #e7f8ef;
  color: #147a42;
}

.status.bad {
  background: #fff1f1;
  color: #bf2f2f;
}

.mini,
.code {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
}

.mini {
  font-size: 12px;
}

.filter select {
  height: 34px;
  padding: 0 10px;
  border-radius: 8px;
  border: 1px solid #c9d5e6;
  background: #fff;
}

.empty {
  margin-top: 12px;
  color: #7b879b;
}

.capture-card {
  margin-top: 12px;
  padding: 14px;
}

.capture-head,
.capture-meta {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}

.capture-head {
  font-weight: 700;
  color: #223a5c;
}

.capture-meta {
  margin-top: 8px;
  color: #63728c;
  font-size: 13px;
}

.code-title {
  margin-top: 12px;
  margin-bottom: 6px;
  color: #223a5c;
  font-weight: 700;
}

.code {
  padding: 12px;
  background: #f8fbff;
  border: 1px solid #e0e8f5;
  border-radius: 8px;
  font-size: 12px;
  line-height: 1.6;
}

@media (max-width: 900px) {
  .page {
    padding: 12px;
  }

  .header,
  .panel-head {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>
