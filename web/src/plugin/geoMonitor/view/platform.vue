<template>
  <div>
    <div class="gva-search-box">
      <el-form
        ref="elSearchFormRef"
        :inline="true"
        :model="searchInfo"
        class="demo-form-inline"
        @keyup.enter="onSubmit"
      >
        <el-form-item label="关键词">
          <el-input
            v-model="searchInfo.keyword"
            placeholder="平台名称或Code"
            clearable
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-select
            v-model="searchInfo.status"
            placeholder="请选择"
            clearable
            style="width: 120px"
          >
            <el-option label="启用" :value="1" />
            <el-option label="停用" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" icon="search" @click="onSubmit">
            查询
          </el-button>
          <el-button icon="refresh" @click="onReset"> 重置 </el-button>
        </el-form-item>
      </el-form>
    </div>
    <div class="gva-table-box">
      <div class="gva-btn-list">
        <el-button type="primary" icon="plus" @click="openDialog">
          新增
        </el-button>
        <el-button
          type="success"
          icon="connection"
          :loading="testingAll"
          style="margin-left: 10px"
          @click="testAll"
        >
          一键测试所有
        </el-button>
      </div>
      <el-table
        ref="multipleTable"
        style="width: 100%"
        tooltip-effect="dark"
        :data="tableData"
        row-key="ID"
      >
        <el-table-column align="left" label="平台名称" min-width="150">
          <template #default="scope">
            <div class="flex items-center gap-2">
              <Icon
                :icon="getPlatformIcon(scope.row.code)"
                width="22"
                height="22"
                :color="getPlatformColor(scope.row.code)"
              />
              <span>{{ scope.row.name }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column align="left" label="Code" prop="code" width="100" />
        <el-table-column align="left" label="API 地址" prop="apiBase" min-width="240" show-overflow-tooltip />
        <el-table-column align="left" label="状态" width="80">
          <template #default="scope">
            <el-tag :type="scope.row.status === 1 ? 'success' : 'danger'" size="small">
              {{ scope.row.status === 1 ? '启用' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column align="left" label="连通状态" width="110">
          <template #default="scope">
            <template v-if="testStatusMap[scope.row.ID]">
              <el-tag
                v-if="testStatusMap[scope.row.ID].status === 'connected'"
                type="success"
                size="small"
              >
                已连接
              </el-tag>
              <el-tag
                v-else-if="testStatusMap[scope.row.ID].status === 'failed'"
                type="danger"
                size="small"
              >
                连接失败
              </el-tag>
              <el-tag
                v-else-if="testStatusMap[scope.row.ID].status === 'unconfigured'"
                type="warning"
                size="small"
              >
                未配置Key
              </el-tag>
            </template>
            <template v-else>
              <span class="text-gray-400 text-sm">未检测</span>
            </template>
          </template>
        </el-table-column>
        <el-table-column align="left" label="排序" prop="sort" width="60" />
        <el-table-column align="left" label="创建时间" width="170">
          <template #default="scope">
            {{ formatDate(scope.row.CreatedAt) }}
          </template>
        </el-table-column>
        <el-table-column align="left" label="操作" fixed="right" min-width="260">
          <template #default="scope">
            <el-button
              type="primary"
              link
              icon="connection"
              class="table-button"
              @click="testConnectivity(scope.row)"
            >
              测试
            </el-button>
            <el-button
              type="primary"
              link
              icon="edit"
              class="table-button"
              @click="updatePlatformFunc(scope.row)"
            >
              编辑
            </el-button>
            <el-button
              type="primary"
              link
              icon="delete"
              @click="deleteRow(scope.row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="gva-pagination">
        <el-pagination
          layout="total, sizes, prev, pager, next, jumper"
          :current-page="page"
          :page-size="pageSize"
          :page-sizes="[10, 30, 50, 100]"
          :total="total"
          @current-change="handleCurrentChange"
          @size-change="handleSizeChange"
        />
      </div>
    </div>

    <el-drawer
      v-model="dialogFormVisible"
      destroy-on-close
      size="600"
      :show-close="false"
      :before-close="closeDialog"
    >
      <template #header>
        <div class="flex justify-between items-center">
          <span class="text-lg">{{ type === 'create' ? '新增平台' : '编辑平台' }}</span>
          <div>
            <el-button type="primary" @click="enterDialog"> 确 定 </el-button>
            <el-button @click="closeDialog"> 取 消 </el-button>
          </div>
        </div>
      </template>

      <el-form
        ref="elFormRef"
        :model="formData"
        label-position="top"
        :rules="rule"
        label-width="80px"
      >
        <el-form-item label="平台名称:" prop="name">
          <el-input
            v-model="formData.name"
            :clearable="true"
            placeholder="如：DeepSeek"
          />
        </el-form-item>
        <el-form-item label="平台渠道:" prop="code">
          <el-select
            v-model="formData.code"
            placeholder="请选择平台渠道"
            style="width: 100%"
            :disabled="type === 'update'"
            @change="onCodeChange"
          >
            <el-option
              v-for="opt in platformOptions"
              :key="opt.code"
              :label="opt.name"
              :value="opt.code"
            >
              <div class="flex items-center gap-2">
                <Icon :icon="opt.icon" width="20" height="20" :color="opt.color" />
                <span>{{ opt.name }}</span>
                <span class="text-gray-400 text-xs ml-auto">{{ opt.code }}</span>
              </div>
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="API 地址:" prop="apiBase">
          <el-input
            v-model="formData.apiBase"
            :clearable="true"
            placeholder="如：https://api.deepseek.com"
          />
        </el-form-item>
        <el-form-item label="API Key:">
          <el-input
            v-model="formData.apiKey"
            type="password"
            show-password
            :clearable="true"
            placeholder="请输入 API Key"
          />
        </el-form-item>
        <el-form-item label="状态:" prop="status">
          <el-switch
            v-model="formData.status"
            :active-value="1"
            :inactive-value="0"
            active-text="启用"
            inactive-text="停用"
          />
        </el-form-item>
        <el-form-item label="排序:" prop="sort">
          <el-input-number
            v-model="formData.sort"
            :min="0"
            :max="999"
          />
        </el-form-item>
        <el-form-item label="备注:">
          <el-input
            v-model="formData.remark"
            type="textarea"
            :rows="3"
            :clearable="true"
            placeholder="备注信息"
          />
        </el-form-item>
      </el-form>
    </el-drawer>
  </div>
</template>

<script setup>
  import {
    getPlatformList,
    getPlatform,
    createPlatform,
    updatePlatform,
    deletePlatform,
    testPlatform,
    testAllPlatforms
  } from '@/plugin/geoMonitor/api/platform'
  import { platformOptions, getPlatformByCode } from '@/plugin/geoMonitor/utils/platformConfig'
  import { formatDate } from '@/utils/format'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { Icon } from '@iconify/vue'
  import { ref, reactive } from 'vue'

  defineOptions({
    name: 'GeoMonitorPlatform'
  })

  const formData = ref({
    name: '',
    code: '',
    apiBase: '',
    apiKey: '',
    status: 1,
    sort: 0,
    remark: ''
  })

  const rule = reactive({
    name: [{ required: true, message: '请输入平台名称', trigger: 'blur' }],
    code: [{ required: true, message: '请选择平台渠道', trigger: 'change' }],
    apiBase: [{ required: true, message: '请输入API地址', trigger: 'blur' }]
  })

  const elFormRef = ref()
  const elSearchFormRef = ref()

  const page = ref(1)
  const total = ref(0)
  const pageSize = ref(10)
  const tableData = ref([])
  const searchInfo = ref({})

  const onReset = () => {
    searchInfo.value = {}
    getTableData()
  }

  const onSubmit = () => {
    page.value = 1
    getTableData()
  }

  const handleSizeChange = (val) => {
    pageSize.value = val
    getTableData()
  }

  const handleCurrentChange = (val) => {
    page.value = val
    getTableData()
  }

  const getTableData = async () => {
    const params = {
      page: page.value,
      pageSize: pageSize.value
    }
    if (searchInfo.value.keyword) {
      params.name = searchInfo.value.keyword
      params.code = searchInfo.value.keyword
    }
    if (searchInfo.value.status !== undefined && searchInfo.value.status !== null && searchInfo.value.status !== '') {
      params.status = searchInfo.value.status
    }
    const table = await getPlatformList(params)
    if (table.code === 0) {
      tableData.value = table.data.list
      total.value = table.data.total
      page.value = table.data.page
      pageSize.value = table.data.pageSize
    }
  }

  getTableData()

  // 新增/编辑弹窗
  const type = ref('')
  const dialogFormVisible = ref(false)

  const openDialog = () => {
    type.value = 'create'
    formData.value = {
      name: '',
      code: '',
      apiBase: '',
      apiKey: '',
      status: 1,
      sort: 0,
      remark: ''
    }
    dialogFormVisible.value = true
  }

  const closeDialog = () => {
    dialogFormVisible.value = false
  }

  const onCodeChange = (code) => {
    const preset = getPlatformByCode(code)
    if (preset) {
      formData.value.name = preset.name
      formData.value.apiBase = preset.apiBase
    }
  }

  const updatePlatformFunc = async (row) => {
    const res = await getPlatform(row.ID)
    type.value = 'update'
    if (res.code === 0) {
      formData.value = res.data
      dialogFormVisible.value = true
    }
  }

  const enterDialog = async () => {
    elFormRef.value?.validate(async (valid) => {
      if (!valid) return
      let res
      switch (type.value) {
        case 'create':
          res = await createPlatform(formData.value)
          break
        case 'update':
          res = await updatePlatform(formData.value.ID, formData.value)
          break
        default:
          res = await createPlatform(formData.value)
          break
      }
      if (res.code === 0) {
        ElMessage({ type: 'success', message: '操作成功' })
        closeDialog()
        getTableData()
      }
    })
  }

  const deleteRow = (row) => {
    ElMessageBox.confirm('确定要删除该平台吗?', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }).then(async () => {
      const res = await deletePlatform(row.ID)
      if (res.code === 0) {
        ElMessage({ type: 'success', message: '删除成功' })
        if (tableData.value.length === 1 && page.value > 1) {
          page.value--
        }
        getTableData()
      }
    })
  }

  const getPlatformIcon = (code) => {
    const preset = getPlatformByCode(code)
    return preset ? preset.icon : 'mdi:api'
  }

  const getPlatformColor = (code) => {
    const preset = getPlatformByCode(code)
    return preset ? preset.color : '#999'
  }

  const testStatusMap = ref({})
  const testingAll = ref(false)

  const testConnectivity = async (row) => {
    if (!row.apiKey) {
      ElMessage({ type: 'warning', message: '请先配置 API Key' })
      return
    }
    const res = await testPlatform(row.ID)
    if (res.code === 0) {
      testStatusMap.value[row.ID] = { status: 'connected', message: '连接成功' }
      ElMessage({ type: 'success', message: '连接成功！API Key 有效' })
    } else {
      testStatusMap.value[row.ID] = { status: 'failed', message: res.msg }
      ElMessage({ type: 'error', message: res.msg || '连接失败' })
    }
  }

  const testAll = async () => {
    testingAll.value = true
    const res = await testAllPlatforms()
    if (res.code === 0) {
      const map = {}
      res.data.forEach((item) => {
        map[item.id] = { status: item.status, message: item.message }
      })
      testStatusMap.value = map
      const connected = res.data.filter((i) => i.status === 'connected').length
      const failed = res.data.filter((i) => i.status === 'failed').length
      const unconfigured = res.data.filter((i) => i.status === 'unconfigured').length
      ElMessage({
        type: 'success',
        message: `测试完成：已连接 ${connected} / 失败 ${failed} / 未配置 ${unconfigured}`
      })
    } else {
      ElMessage({ type: 'error', message: res.msg || '测试失败' })
    }
    testingAll.value = false
  }
</script>
