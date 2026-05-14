<template>
  <div>
    <div class="gva-search-box">
      <el-form :model="formData" label-position="top">
        <el-form-item label="监控话题">
          <el-select v-model="formData.topicId" placeholder="请选择话题" style="width: 100%">
            <el-option v-for="item in topicOptions" :key="item.ID" :label="item.name" :value="item.ID" />
          </el-select>
        </el-form-item>
        <el-form-item label="采集模式">
          <el-radio-group v-model="formData.mode" @change="loadPlatforms">
            <el-radio value="api">API</el-radio>
            <el-radio value="playwright">Playwright</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="选择平台">
          <el-select v-model="formData.platformIds" multiple placeholder="请选择平台" style="width: 100%">
            <el-option v-for="item in platformOptions" :key="item.ID" :label="item.name" :value="item.ID" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="running" @click="submitRun">执行采集</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="gva-table-box" v-if="taskId">
      <div class="mb-4">任务ID：{{ taskId }} / 状态：{{ taskStatus }}</div>
      <el-table :data="resultList">
        <el-table-column prop="platformName" label="平台" min-width="140" />
        <el-table-column prop="status" label="状态" width="100" />
        <el-table-column prop="durationMs" label="耗时(ms)" width="120" />
        <el-table-column prop="answer" label="回答摘要" min-width="320" show-overflow-tooltip />
        <el-table-column prop="errorMsg" label="错误" min-width="180" show-overflow-tooltip />
        <el-table-column label="诊断" width="190">
          <template #default="scope">
            <el-button link type="primary" @click="openCitations(scope.row)">引用源</el-button>
            <el-button link type="primary" @click="openRunLog(scope.row)">日志</el-button>
          </template>
        </el-table-column>
        <el-table-column label="截图" width="120">
          <template #default="scope">
            <el-link v-if="scope.row.screenshotPath" :href="normalizeAssetPath(scope.row.screenshotPath)" target="_blank">查看</el-link>
            <span v-else>-</span>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="citationsDialogVisible" title="引用源 / 搜索结果" width="760px">
      <el-empty v-if="citationItems.length === 0" description="暂无引用源" />
      <el-table v-else :data="citationItems" max-height="420">
        <el-table-column prop="index" label="#" width="60" />
        <el-table-column prop="title" label="标题" min-width="160" show-overflow-tooltip />
        <el-table-column prop="snippet" label="摘要" min-width="220" show-overflow-tooltip />
        <el-table-column label="链接" min-width="180">
          <template #default="scope">
            <el-link v-if="scope.row.url" :href="scope.row.url" target="_blank">{{ scope.row.url }}</el-link>
            <span v-else>-</span>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <el-dialog v-model="runLogDialogVisible" title="运行日志" width="760px">
      <el-empty v-if="runLogItems.length === 0" description="暂无日志" />
      <el-timeline v-else>
        <el-timeline-item v-for="(item, index) in runLogItems" :key="index" :timestamp="item.time">
          <div>{{ item.step }} / {{ item.status }} / {{ item.durationMs }}ms</div>
          <div>{{ item.message }}</div>
        </el-timeline-item>
      </el-timeline>
    </el-dialog>
  </div>
</template>

<script setup>
  import { ref, onMounted } from 'vue'
  import { ElMessage } from 'element-plus'
  import { getAllTopics } from '@/plugin/geoMonitor/api/topic'
  import { getEnabledPlatformsByMode } from '@/plugin/geoMonitor/api/platform'
  import { runCollection, getCollectionResultList } from '@/plugin/geoMonitor/api/collect'

  defineOptions({ name: 'GeoMonitorCollect' })

  const formData = ref({ topicId: '', mode: 'api', platformIds: [] })
  const topicOptions = ref([])
  const platformOptions = ref([])
  const resultList = ref([])
  const taskId = ref(0)
  const taskStatus = ref('')
  const running = ref(false)
  const citationsDialogVisible = ref(false)
  const runLogDialogVisible = ref(false)
  const citationItems = ref([])
  const runLogItems = ref([])

  const normalizeAssetPath = (path) => {
    if (!path) return ''
    if (path.startsWith('http') || path.startsWith('/')) return path
    return `/${path}`
  }

  const parseJsonArray = (value) => {
    if (!value) return []
    try {
      const data = JSON.parse(value)
      return Array.isArray(data) ? data : []
    } catch (e) {
      return []
    }
  }

  const openCitations = (row) => {
    citationItems.value = parseJsonArray(row.citations)
    citationsDialogVisible.value = true
  }

  const openRunLog = (row) => {
    runLogItems.value = parseJsonArray(row.runLog)
    runLogDialogVisible.value = true
  }

  const loadTopics = async () => {
    const res = await getAllTopics()
    if (res.code === 0) topicOptions.value = res.data.list
  }

  const loadPlatforms = async () => {
    formData.value.platformIds = []
    const res = await getEnabledPlatformsByMode(formData.value.mode)
    if (res.code === 0) platformOptions.value = res.data.list
  }

  const submitRun = async () => {
    if (!formData.value.topicId || formData.value.platformIds.length === 0) {
      ElMessage({ type: 'warning', message: '请先选择话题和平台' })
      return
    }
    running.value = true
    const res = await runCollection(formData.value)
    running.value = false
    if (res.code !== 0) return
    taskId.value = res.data.taskId
    taskStatus.value = res.data.status
    await loadResults()
    ElMessage({ type: 'success', message: '执行完成' })
  }

  const loadResults = async () => {
    const res = await getCollectionResultList({ taskId: taskId.value, page: 1, pageSize: 100 })
    if (res.code === 0) resultList.value = res.data.list
  }

  onMounted(async () => {
    await loadTopics()
    await loadPlatforms()
  })
</script>
