<script setup>
import { computed, onMounted, ref, h } from 'vue'
import { NButton, NDataTable, NInput, NInputNumber, NSpace, NSwitch, NTag, useMessage } from 'naive-ui'
import { ApiGetCmdAnalysis, ApiConfirmCmdName, ApiGetIgnoredCmdList, ApiSetCmdIgnore } from '../api'

const nmessage = useMessage()
const loading = ref(false)
const onlyUnknown = ref(true)
const onlyNotIgnored = ref(false)
const onlyIgnored = ref(false)
const rows = ref([])
const checkedRowKeys = ref([])
const confirmMap = ref({})
const cmdIdFilter = ref(null)
const ignoredList = ref([])

const readyCount = computed(() => {
  let n = 0
  for (const row of rows.value) {
    const name = (confirmMap.value[row.cmd_id] || '').trim()
    if (name && name !== '未映射') n++
  }
  return n
})

function fmtTime(unixSec) {
  if (!unixSec) return '-'
  const d = new Date(Number(unixSec) * 1000)
  if (Number.isNaN(d.getTime())) return String(unixSec)
  return d.toLocaleString('zh-CN', { hour12: false })
}

function confidenceType(v) {
  if (v >= 0.9) return 'success'
  if (v >= 0.75) return 'warning'
  return 'default'
}

function setDefaultConfirmName(items) {
  const next = { ...confirmMap.value }
  for (const it of items) {
    if (!next[it.cmd_id]) {
      next[it.cmd_id] = it.cmd_name && it.cmd_name !== '未映射' ? it.cmd_name : (it.suggested_name || '')
    }
  }
  confirmMap.value = next
}

async function loadIgnored() {
  const res = await ApiGetIgnoredCmdList()
  ignoredList.value = res?.data?.data?.items || []
}

async function loadData() {
  loading.value = true
  try {
    const params = {
      only_unknown: onlyUnknown.value ? 1 : 0,
      only_not_ignored: onlyNotIgnored.value ? 1 : 0,
      only_ignored: onlyIgnored.value ? 1 : 0,
      limit: 500,
    }
    if (cmdIdFilter.value) params.cmd_id = Number(cmdIdFilter.value)
    const res = await ApiGetCmdAnalysis(params)
    const items = res?.data?.data?.items || []
    rows.value = items
    setDefaultConfirmName(items)
    await loadIgnored()
  } finally {
    loading.value = false
  }
}

async function toggleIgnore(row) {
  const next = !row.ignored
  const res = await ApiSetCmdIgnore({ cmd_id: row.cmd_id, ignore: next })
  if (res?.data?.code === 200) {
    nmessage.success(`cmd_id=${row.cmd_id} 已${next ? '加入' : '移出'}忽略`) 
    await loadData()
  } else {
    nmessage.error(res?.data?.msg || '忽略设置失败')
  }
}

async function confirmOne(row) {
  const name = (confirmMap.value[row.cmd_id] || '').trim()
  if (!name || name === '未映射') {
    nmessage.error(`cmd_id=${row.cmd_id} 的确认名称无效`)
    return
  }
  const res = await ApiConfirmCmdName({ cmd_id: row.cmd_id, cmd_name: name })
  if (res?.data?.code === 200) {
    nmessage.success(`已确认 cmd_id=${row.cmd_id}`)
    await loadData()
  } else {
    nmessage.error(res?.data?.msg || '确认失败')
  }
}

async function confirmBatch() {
  const selected = rows.value.filter((r) => checkedRowKeys.value.includes(r.cmd_id))
  if (selected.length === 0) {
    nmessage.error('请先勾选要确认的行')
    return
  }
  const items = selected.map((r) => ({ cmd_id: r.cmd_id, cmd_name: (confirmMap.value[r.cmd_id] || '').trim() }))
  const invalid = items.filter((x) => !x.cmd_name || x.cmd_name === '未映射')
  if (invalid.length > 0) {
    nmessage.error(`有 ${invalid.length} 条确认名为空或非法`)
    return
  }
  const res = await ApiConfirmCmdName({ items })
  if (res?.data?.code === 200) {
    const updated = res?.data?.data?.updated || 0
    nmessage.success(`批量确认完成，更新 ${updated} 条`)
    await loadData()
  } else {
    nmessage.error(res?.data?.msg || '批量确认失败')
  }
}

const columns = [
  { type: 'selection' },
  { title: 'cmd_id', key: 'cmd_id', width: 90, sorter: (a, b) => a.cmd_id - b.cmd_id },
  {
    title: '当前语义',
    key: 'cmd_name',
    width: 160,
    render(row) {
      const t = row.cmd_name === '未映射' ? 'warning' : 'success'
      return h(NTag, { type: t, bordered: false }, { default: () => row.cmd_name })
    },
  },
  { title: '自动推断', key: 'inferred_semantics', minWidth: 180 },
  {
    title: '建议名称',
    key: 'confirm_name',
    minWidth: 180,
    render(row) {
      return h(NInput, {
        value: confirmMap.value[row.cmd_id] || '',
        onUpdateValue: (v) => { confirmMap.value[row.cmd_id] = v },
        placeholder: '输入确认名称',
      })
    },
  },
  { title: '请求类型', key: 'request_type', width: 140 },
  {
    title: '忽略',
    key: 'ignored',
    width: 100,
    render(row) {
      return h(NSwitch, {
        value: !!row.ignored,
        'onUpdate:value': () => toggleIgnore(row),
      })
    },
  },
  {
    title: 'API请求',
    key: 'apis',
    minWidth: 220,
    render(row) {
      const arr = Array.isArray(row.apis) ? row.apis : []
      return arr.length ? arr.join(' | ') : '-'
    },
  },
  {
    title: 'API示例(method/route/req/resp)',
    key: 'api_examples',
    minWidth: 360,
    render(row) {
      const arr = Array.isArray(row.api_examples) ? row.api_examples : []
      if (!arr.length) return '-'
      return arr.map((it) => `${it.method} ${it.route}\nreq: ${it.req}\nresp: ${it.resp}`).join('\n---\n')
    },
  },
  {
    title: '置信度',
    key: 'confidence',
    width: 120,
    render(row) {
      return h(NTag, { type: confidenceType(row.confidence), bordered: false }, { default: () => `${Math.round(row.confidence * 100)}%` })
    },
  },
  { title: '出现次数', key: 'seen_count', width: 100, sorter: (a, b) => a.seen_count - b.seen_count },
  {
    title: '依据',
    key: 'basis',
    minWidth: 220,
    render(row) {
      const text = Array.isArray(row.basis) ? row.basis.join(' | ') : ''
      return text || '-'
    },
  },
  {
    title: '更新时间',
    key: 'updated_at',
    width: 180,
    render(row) {
      return fmtTime(row.updated_at)
    },
  },
  {
    title: '操作',
    key: 'actions',
    width: 100,
    render(row) {
      return h(NButton, { size: 'small', type: 'primary', onClick: () => confirmOne(row) }, { default: () => '确认' })
    },
  },
]

onMounted(loadData)
</script>

<template>
  <div class="bikamoeapp">
    <div class="bikamoeapp-content">
      <div style="margin: 16px auto;padding: 16px;">
        <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:12px;">
          <div style="font-size: 18px;font-weight: 600;">CMD语义确认中心</div>
        </div>

        <n-space style="margin-bottom:12px;" align="center">
          <span>仅看未映射</span>
          <n-switch v-model:value="onlyUnknown" @update:value="loadData" />
          <span>仅看未忽略</span>
          <n-switch v-model:value="onlyNotIgnored" @update:value="loadData" />
          <span>仅看已忽略</span>
          <n-switch v-model:value="onlyIgnored" @update:value="loadData" />
          <n-input-number
            v-model:value="cmdIdFilter"
            :min="1"
            placeholder="只填cmd_id"
            style="width: 160px"
            clearable
          />
          <n-button @click="loadData">刷新</n-button>
          <n-button type="primary" @click="confirmBatch">批量确认所选</n-button>
          <span>可确认条数: {{ readyCount }}</span>
          <span>当前忽略: {{ ignoredList.join(', ') || '无' }}</span>
        </n-space>

        <n-data-table
          v-model:checked-row-keys="checkedRowKeys"
          :row-key="(row) => row.cmd_id"
          :columns="columns"
          :data="rows"
          :loading="loading"
          :pagination="{ pageSize: 20 }"
          size="small"
        />
      </div>
    </div>
  </div>
</template>

<style scoped></style>

