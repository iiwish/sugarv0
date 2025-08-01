
<template>
  <div>
    <div class="gva-search-box">
      <el-form ref="elSearchFormRef" :inline="true" :model="searchInfo" class="demo-form-inline" @keyup.enter="onSubmit">

        <template v-if="showAllQuery">
          <!-- 将需要控制显示状态的查询条件添加到此范围内 -->
        </template>

        <el-form-item>
          <el-button type="primary" icon="search" @click="onSubmit">查询</el-button>
          <el-button icon="refresh" @click="onReset">重置</el-button>
          <el-button link type="primary" icon="arrow-down" @click="showAllQuery=true" v-if="!showAllQuery">展开</el-button>
          <el-button link type="primary" icon="arrow-up" @click="showAllQuery=false" v-else>收起</el-button>
        </el-form-item>
      </el-form>
    </div>
    <div class="gva-table-box">
        <div class="gva-btn-list">
            <el-button  type="primary" icon="plus" @click="openDialog()">新增</el-button>
            <el-button  icon="delete" style="margin-left: 10px;" :disabled="!multipleSelection.length" @click="onDelete">删除</el-button>
            
        </div>
        <el-table
        ref="multipleTable"
        style="width: 100%"
        tooltip-effect="dark"
        :data="tableData"
        row-key="id"
        @selection-change="handleSelectionChange"
        >
        <el-table-column type="selection" width="55" />
        
            <el-table-column align="left" label="id字段" prop="id" width="120" />

            <el-table-column align="left" label="logType字段" prop="logType" width="120" />

            <el-table-column align="left" label="workspaceId字段" prop="workspaceId" width="120" />

            <el-table-column align="left" label="userId字段" prop="userId" width="120" />

            <el-table-column align="left" label="connectionId字段" prop="connectionId" width="120" />

            <el-table-column align="left" label="agentId字段" prop="agentId" width="120" />

            <el-table-column label="inputPayload字段" prop="inputPayload" width="200">
    <template #default="scope">
        [JSON]
    </template>
</el-table-column>
            <el-table-column align="left" label="status字段" prop="status" width="120" />

            <el-table-column align="left" label="resultSummary字段" prop="resultSummary" width="120" />

            <el-table-column align="left" label="durationMs字段" prop="durationMs" width="120" />

            <el-table-column align="left" label="executedAt字段" prop="executedAt" width="180">
   <template #default="scope">{{ formatDate(scope.row.executedAt) }}</template>
</el-table-column>
        <el-table-column align="left" label="操作" fixed="right" :min-width="appStore.operateMinWith">
            <template #default="scope">
            <el-button  type="primary" link class="table-button" @click="getDetails(scope.row)"><el-icon style="margin-right: 5px"><InfoFilled /></el-icon>查看</el-button>
            <el-button  type="primary" link icon="edit" class="table-button" @click="updateSugarExecutionLogsFunc(scope.row)">编辑</el-button>
            <el-button   type="primary" link icon="delete" @click="deleteRow(scope.row)">删除</el-button>
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
    <el-drawer destroy-on-close :size="appStore.drawerSize" v-model="dialogFormVisible" :show-close="false" :before-close="closeDialog">
       <template #header>
              <div class="flex justify-between items-center">
                <span class="text-lg">{{type==='create'?'新增':'编辑'}}</span>
                <div>
                  <el-button :loading="btnLoading" type="primary" @click="enterDialog">确 定</el-button>
                  <el-button @click="closeDialog">取 消</el-button>
                </div>
              </div>
            </template>

          <el-form :model="formData" label-position="top" ref="elFormRef" :rules="rule" label-width="80px">
            <el-form-item label="id字段:" prop="id">
    <el-input v-model.number="formData.id" :clearable="true" placeholder="请输入id字段" />
</el-form-item>
            <el-form-item label="logType字段:" prop="logType">
    <el-select v-model="formData.logType" placeholder="请选择logType字段" style="width:100%" filterable :clearable="true">
       <el-option v-for="item in ['']" :key="item" :label="item" :value="item" />
    </el-select>
</el-form-item>
            <el-form-item label="workspaceId字段:" prop="workspaceId">
    <el-input v-model="formData.workspaceId" :clearable="true" placeholder="请输入workspaceId字段" />
</el-form-item>
            <el-form-item label="userId字段:" prop="userId">
    <el-input v-model="formData.userId" :clearable="true" placeholder="请输入userId字段" />
</el-form-item>
            <el-form-item label="connectionId字段:" prop="connectionId">
    <el-input v-model="formData.connectionId" :clearable="true" placeholder="请输入connectionId字段" />
</el-form-item>
            <el-form-item label="agentId字段:" prop="agentId">
    <el-input v-model="formData.agentId" :clearable="true" placeholder="请输入agentId字段" />
</el-form-item>
            <el-form-item label="inputPayload字段:" prop="inputPayload">
    // 此字段为json结构，可以前端自行控制展示和数据绑定模式 需绑定json的key为 formData.inputPayload 后端会按照json的类型进行存取
    {{ formData.inputPayload }}
</el-form-item>
            <el-form-item label="status字段:" prop="status">
    <el-select v-model="formData.status" placeholder="请选择status字段" style="width:100%" filterable :clearable="true">
       <el-option v-for="item in ['']" :key="item" :label="item" :value="item" />
    </el-select>
</el-form-item>
            <el-form-item label="resultSummary字段:" prop="resultSummary">
    <el-input v-model="formData.resultSummary" :clearable="true" placeholder="请输入resultSummary字段" />
</el-form-item>
            <el-form-item label="durationMs字段:" prop="durationMs">
    <el-input v-model.number="formData.durationMs" :clearable="true" placeholder="请输入durationMs字段" />
</el-form-item>
            <el-form-item label="executedAt字段:" prop="executedAt">
    <el-date-picker v-model="formData.executedAt" type="date" style="width:100%" placeholder="选择日期" :clearable="true" />
</el-form-item>
          </el-form>
    </el-drawer>

    <el-drawer destroy-on-close :size="appStore.drawerSize" v-model="detailShow" :show-close="true" :before-close="closeDetailShow" title="查看">
            <el-descriptions :column="1" border>
                    <el-descriptions-item label="id字段">
    {{ detailForm.id }}
</el-descriptions-item>
                    <el-descriptions-item label="logType字段">
    {{ detailForm.logType }}
</el-descriptions-item>
                    <el-descriptions-item label="workspaceId字段">
    {{ detailForm.workspaceId }}
</el-descriptions-item>
                    <el-descriptions-item label="userId字段">
    {{ detailForm.userId }}
</el-descriptions-item>
                    <el-descriptions-item label="connectionId字段">
    {{ detailForm.connectionId }}
</el-descriptions-item>
                    <el-descriptions-item label="agentId字段">
    {{ detailForm.agentId }}
</el-descriptions-item>
                    <el-descriptions-item label="inputPayload字段">
    {{ detailForm.inputPayload }}
</el-descriptions-item>
                    <el-descriptions-item label="status字段">
    {{ detailForm.status }}
</el-descriptions-item>
                    <el-descriptions-item label="resultSummary字段">
    {{ detailForm.resultSummary }}
</el-descriptions-item>
                    <el-descriptions-item label="durationMs字段">
    {{ detailForm.durationMs }}
</el-descriptions-item>
                    <el-descriptions-item label="executedAt字段">
    {{ detailForm.executedAt }}
</el-descriptions-item>
            </el-descriptions>
        </el-drawer>

  </div>
</template>

<script setup>
import {
  createSugarExecutionLogs,
  deleteSugarExecutionLogs,
  deleteSugarExecutionLogsByIds,
  updateSugarExecutionLogs,
  findSugarExecutionLogs,
  getSugarExecutionLogsList
} from '@/api/sugar/sugarExecutionLogs'

// 全量引入格式化工具 请按需保留
import { getDictFunc, formatDate, formatBoolean, filterDict ,filterDataSource, returnArrImg, onDownloadFile } from '@/utils/format'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ref, reactive } from 'vue'
import { useAppStore } from "@/pinia"




defineOptions({
    name: 'SugarExecutionLogs'
})

// 提交按钮loading
const btnLoading = ref(false)
const appStore = useAppStore()

// 控制更多查询条件显示/隐藏状态
const showAllQuery = ref(false)

// 自动化生成的字典（可能为空）以及字段
const formData = ref({
            id: undefined,
            logType: null,
            workspaceId: '',
            userId: '',
            connectionId: '',
            agentId: '',
            inputPayload: {},
            status: null,
            resultSummary: '',
            durationMs: undefined,
            executedAt: new Date(),
        })



// 验证规则
const rule = reactive({
})

const elFormRef = ref()
const elSearchFormRef = ref()

// =========== 表格控制部分 ===========
const page = ref(1)
const total = ref(0)
const pageSize = ref(10)
const tableData = ref([])
const searchInfo = ref({})
// 重置
const onReset = () => {
  searchInfo.value = {}
  getTableData()
}

// 搜索
const onSubmit = () => {
  elSearchFormRef.value?.validate(async(valid) => {
    if (!valid) return
    page.value = 1
    getTableData()
  })
}

// 分页
const handleSizeChange = (val) => {
  pageSize.value = val
  getTableData()
}

// 修改页面容量
const handleCurrentChange = (val) => {
  page.value = val
  getTableData()
}

// 查询
const getTableData = async() => {
  const table = await getSugarExecutionLogsList({ page: page.value, pageSize: pageSize.value, ...searchInfo.value })
  if (table.code === 0) {
    tableData.value = table.data.list
    total.value = table.data.total
    page.value = table.data.page
    pageSize.value = table.data.pageSize
  }
}

getTableData()

// ============== 表格控制部分结束 ===============

// 获取需要的字典 可能为空 按需保留
const setOptions = async () =>{
}

// 获取需要的字典 可能为空 按需保留
setOptions()


// 多选数据
const multipleSelection = ref([])
// 多选
const handleSelectionChange = (val) => {
    multipleSelection.value = val
}

// 删除行
const deleteRow = (row) => {
    ElMessageBox.confirm('确定要删除吗?', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
    }).then(() => {
            deleteSugarExecutionLogsFunc(row)
        })
    }

// 多选删除
const onDelete = async() => {
  ElMessageBox.confirm('确定要删除吗?', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async() => {
      const ids = []
      if (multipleSelection.value.length === 0) {
        ElMessage({
          type: 'warning',
          message: '请选择要删除的数据'
        })
        return
      }
      multipleSelection.value &&
        multipleSelection.value.map(item => {
          ids.push(item.id)
        })
      const res = await deleteSugarExecutionLogsByIds({ ids })
      if (res.code === 0) {
        ElMessage({
          type: 'success',
          message: '删除成功'
        })
        if (tableData.value.length === ids.length && page.value > 1) {
          page.value--
        }
        getTableData()
      }
      })
    }

// 行为控制标记（弹窗内部需要增还是改）
const type = ref('')

// 更新行
const updateSugarExecutionLogsFunc = async(row) => {
    const res = await findSugarExecutionLogs({ id: row.id })
    type.value = 'update'
    if (res.code === 0) {
        formData.value = res.data
        dialogFormVisible.value = true
    }
}


// 删除行
const deleteSugarExecutionLogsFunc = async (row) => {
    const res = await deleteSugarExecutionLogs({ id: row.id })
    if (res.code === 0) {
        ElMessage({
                type: 'success',
                message: '删除成功'
            })
            if (tableData.value.length === 1 && page.value > 1) {
            page.value--
        }
        getTableData()
    }
}

// 弹窗控制标记
const dialogFormVisible = ref(false)

// 打开弹窗
const openDialog = () => {
    type.value = 'create'
    dialogFormVisible.value = true
}

// 关闭弹窗
const closeDialog = () => {
    dialogFormVisible.value = false
    formData.value = {
        id: undefined,
        logType: null,
        workspaceId: '',
        userId: '',
        connectionId: '',
        agentId: '',
        inputPayload: {},
        status: null,
        resultSummary: '',
        durationMs: undefined,
        executedAt: new Date(),
        }
}
// 弹窗确定
const enterDialog = async () => {
     btnLoading.value = true
     elFormRef.value?.validate( async (valid) => {
             if (!valid) return btnLoading.value = false
              let res
              switch (type.value) {
                case 'create':
                  res = await createSugarExecutionLogs(formData.value)
                  break
                case 'update':
                  res = await updateSugarExecutionLogs(formData.value)
                  break
                default:
                  res = await createSugarExecutionLogs(formData.value)
                  break
              }
              btnLoading.value = false
              if (res.code === 0) {
                ElMessage({
                  type: 'success',
                  message: '创建/更改成功'
                })
                closeDialog()
                getTableData()
              }
      })
}

const detailForm = ref({})

// 查看详情控制标记
const detailShow = ref(false)


// 打开详情弹窗
const openDetailShow = () => {
  detailShow.value = true
}


// 打开详情
const getDetails = async (row) => {
  // 打开弹窗
  const res = await findSugarExecutionLogs({ id: row.id })
  if (res.code === 0) {
    detailForm.value = res.data
    openDetailShow()
  }
}


// 关闭详情弹窗
const closeDetailShow = () => {
  detailShow.value = false
  detailForm.value = {}
}


</script>

<style>

</style>
