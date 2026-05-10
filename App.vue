<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NMessageProvider, NDialogProvider, NConfigProvider } from 'naive-ui'
import { zhCN, dateZhCN } from 'naive-ui'
import { ChevronLeft } from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()

const isHome = computed(() => route.path === '/')

const pageTitles = {
  '/teamuser':         '同盟成员',
  '/task':             '攻城任务',
  '/groupWu':          '分组武勋',
  '/battle':           '战报数据',
  '/battle/team':      '队伍查询',
  '/leaderboard':      '排行榜看板',
  '/team':             '队伍深度分析',
  '/battle/analysis':  '战报聚合分析',
  '/cmd-review':       'CMD 语义确认',
  '/data-readiness':   '业务可得性',
  '/member-diagnosis': '成员抓取诊断',
  '/traffic-apis':     '流量接口目录',
}

const pageTitle = computed(() => pageTitles[route.path] || '')
</script>

<template>
  <n-config-provider :locale="zhCN" :date-locale="dateZhCN">
  <n-dialog-provider>
  <n-message-provider>

    <header v-if="!isHome" class="g-nav">
      <button class="g-nav-back" @click="router.push('/')">
        <ChevronLeft :size="15" />
        首页
      </button>
      <span class="g-nav-title">{{ pageTitle }}</span>
    </header>

    <router-view v-slot="{ Component, route: r }">
      <keep-alive include="Index">
        <component :is="Component" :key="r.path" />
      </keep-alive>
    </router-view>

  </n-message-provider>
  </n-dialog-provider>
  </n-config-provider>
</template>

<style>
.g-nav {
  position: sticky;
  top: 0;
  z-index: 200;
  display: flex;
  align-items: center;
  gap: 12px;
  height: 48px;
  padding: 0 20px;
  background: #fff;
  border-bottom: 1px solid #e2e8f0;
  box-shadow: 0 1px 3px rgba(0,0,0,0.04);
}

.g-nav-back {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  border: 1px solid #e2e8f0;
  border-radius: 7px;
  background: #fff;
  font-size: 13px;
  font-weight: 500;
  color: #475569;
  cursor: pointer;
  transition: border-color 0.15s, color 0.15s;
}

.g-nav-back:hover {
  border-color: #a5b4fc;
  color: #4f46e5;
}

.g-nav-title {
  font-size: 14px;
  font-weight: 600;
  color: #1e293b;
}
</style>
