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
            v-model="searchInfo.name"
            placeholder="话题名称"
            clearable
          />
        </el-form-item>
        <el-form-item label="监控类型">
          <el-select
            v-model="searchInfo.type"
            placeholder="请选择"
            clearable
            style="width: 140px"
          >
            <el-option
              v-for="opt in typeOptions"
              :key="opt.value"
              :label="opt.label"
              :value="opt.value"
            />
          </el-select>
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
      </div>
      <el-table
        ref="multipleTable"
        style="width: 100%"
        tooltip-effect="dark"
        :data="tableData"
        row-key="ID"
      >
        <el-table-column align="left" label="序号" type="index" width="60" />
        <el-table-column align="left" label="监控类型" width="110">
          <template #default="scope">
            <el-tag size="small">{{ getTypeLabel(scope.row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column align="left" label="话题名称" prop="name" min-width="150" />
        <el-table-column align="left" label="AI搜索提示词" prop="prompt" min-width="300" show-overflow-tooltip />
        <el-table-column align="left" label="创建人" prop="userName" width="120" />
        <el-table-column align="left" label="状态" width="80">
          <template #default="scope">
            <el-tag :type="scope.row.status === 1 ? 'success' : 'danger'" size="small">
              {{ scope.row.status === 1 ? '启用' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column align="left" label="创建时间" width="170">
          <template #default="scope">
            {{ formatDate(scope.row.CreatedAt) }}
          </template>
        </el-table-column>
        <el-table-column align="left" label="操作" fixed="right" min-width="160">
          <template #default="scope">
            <el-button
              type="primary"
              link
              icon="edit"
              class="table-button"
              @click="updateTopicFunc(scope.row)"
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
          <span class="text-lg">{{ type === 'create' ? '新增话题' : '编辑话题' }}</span>
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
        <el-form-item label="监控类型:" prop="type">
          <el-select
            v-model="formData.type"
            placeholder="请选择监控类型"
            style="width: 100%"
          >
            <el-option
              v-for="opt in typeOptions"
              :key="opt.value"
              :label="opt.label"
              :value="opt.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="话题名称:" prop="name">
          <el-input
            v-model="formData.name"
            :clearable="true"
            placeholder="如：AI回答准确性测试"
          />
        </el-form-item>
        <el-form-item label="AI搜索提示词:" prop="prompt">
          <el-input
            v-model="formData.prompt"
            type="textarea"
            :rows="6"
            :clearable="true"
            placeholder="请输入发送给 AI 的提示词"
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
    getTopicList,
    getTopic,
    createTopic,
    updateTopic,
    deleteTopic
  } from '@/plugin/geoMonitor/api/topic'
  import { formatDate } from '@/utils/format'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { ref, reactive } from 'vue'

  defineOptions({
    name: 'GeoMonitorTopic'
  })

  const typeOptions = [
    { label: '事实核查', value: 'fact_check' },
    { label: '实时资讯', value: 'news' },
    { label: '知识问答', value: 'knowledge' },
    { label: '政策解读', value: 'policy' },
    { label: '技术评测', value: 'tech' },
    { label: '其他', value: 'other' }
  ]

  function getTypeLabel(value) {
    const opt = typeOptions.find((o) => o.value === value)
    return opt ? opt.label : value
  }

  const formData = ref({
    type: '',
    name: '',
    prompt: '',
    status: 1,
    remark: ''
  })

  const rule = reactive({
    type: [{ required: true, message: '请选择监控类型', trigger: 'change' }],
    name: [{ required: true, message: '请输入话题名称', trigger: 'blur' }],
    prompt: [{ required: true, message: '请输入AI搜索提示词', trigger: 'blur' }]
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
    if (searchInfo.value.name) {
      params.name = searchInfo.value.name
    }
    if (searchInfo.value.type) {
      params.type = searchInfo.value.type
    }
    if (searchInfo.value.status !== undefined && searchInfo.value.status !== null && searchInfo.value.status !== '') {
      params.status = searchInfo.value.status
    }
    const table = await getTopicList(params)
    if (table.code === 0) {
      tableData.value = table.data.list
      total.value = table.data.total
      page.value = table.data.page
      pageSize.value = table.data.pageSize
    }
  }

  getTableData()

  const type = ref('')
  const dialogFormVisible = ref(false)

  const openDialog = () => {
    type.value = 'create'
    formData.value = {
      type: '',
      name: '',
      prompt: '',
      status: 1,
      remark: ''
    }
    dialogFormVisible.value = true
  }

  const closeDialog = () => {
    dialogFormVisible.value = false
  }

  const updateTopicFunc = async (row) => {
    const res = await getTopic(row.ID)
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
          res = await createTopic(formData.value)
          break
        case 'update':
          res = await updateTopic(formData.value.ID, formData.value)
          break
        default:
          res = await createTopic(formData.value)
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
    ElMessageBox.confirm('确定要删除该话题吗?', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }).then(async () => {
      const res = await deleteTopic(row.ID)
      if (res.code === 0) {
        ElMessage({ type: 'success', message: '删除成功' })
        if (tableData.value.length === 1 && page.value > 1) {
          page.value--
        }
        getTableData()
      }
    })
  }
</script>
