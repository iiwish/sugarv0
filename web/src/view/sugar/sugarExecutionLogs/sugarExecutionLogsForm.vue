
<template>
  <div>
    <div class="gva-form-box">
      <el-form :model="formData" ref="elFormRef" label-position="right" :rules="rule" label-width="80px">
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
        <el-form-item>
          <el-button :loading="btnLoading" type="primary" @click="save">保存</el-button>
          <el-button type="primary" @click="back">返回</el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup>
import {
  createSugarExecutionLogs,
  updateSugarExecutionLogs,
  findSugarExecutionLogs
} from '@/api/sugar/sugarExecutionLogs'

defineOptions({
    name: 'SugarExecutionLogsForm'
})

// 自动获取字典
import { getDictFunc } from '@/utils/format'
import { useRoute, useRouter } from "vue-router"
import { ElMessage } from 'element-plus'
import { ref, reactive } from 'vue'


const route = useRoute()
const router = useRouter()

// 提交按钮loading
const btnLoading = ref(false)

const type = ref('')
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

// 初始化方法
const init = async () => {
 // 建议通过url传参获取目标数据ID 调用 find方法进行查询数据操作 从而决定本页面是create还是update 以下为id作为url参数示例
    if (route.query.id) {
      const res = await findSugarExecutionLogs({ ID: route.query.id })
      if (res.code === 0) {
        formData.value = res.data
        type.value = 'update'
      }
    } else {
      type.value = 'create'
    }
}

init()
// 保存按钮
const save = async() => {
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
           }
       })
}

// 返回按钮
const back = () => {
    router.go(-1)
}

</script>

<style>
</style>
