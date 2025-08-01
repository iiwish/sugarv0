
<template>
  <div>
    <div class="gva-form-box">
      <el-form :model="formData" ref="elFormRef" label-position="right" :rules="rule" label-width="80px">
        <el-form-item label="name字段:" prop="name">
    <el-input v-model="formData.name" :clearable="true" placeholder="请输入name字段" />
</el-form-item>
        <el-form-item label="description字段:" prop="description">
    <el-input v-model="formData.description" :clearable="true" placeholder="请输入description字段" />
</el-form-item>
        <el-form-item label="系统预置, 团队自定义:" prop="agentType">
    <el-select v-model="formData.agentType" placeholder="请选择系统预置, 团队自定义" style="width:100%" filterable :clearable="true">
       <el-option v-for="item in ['']" :key="item" :label="item" :value="item" />
    </el-select>
</el-form-item>
        <el-form-item label="teamId字段:" prop="teamId">
    <el-input v-model="formData.teamId" :clearable="true" placeholder="请输入teamId字段" />
</el-form-item>
        <el-form-item label="定义 Agent 的调用方式, 如 API URL, headers 等:" prop="endpointConfig">
    <el-select v-model="formData.endpointConfig" placeholder="请选择定义 Agent 的调用方式, 如 API URL, headers 等" style="width:100%" filterable :clearable="true">
       <el-option v-for="item in ['']" :key="item" :label="item" :value="item" />
    </el-select>
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
  createSugarAgents,
  updateSugarAgents,
  findSugarAgents
} from '@/api/sugar/sugarAgents'

defineOptions({
    name: 'SugarAgentsForm'
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
            name: '',
            description: '',
            agentType: null,
            teamId: '',
            endpointConfig: null,
        })
// 验证规则
const rule = reactive({
})

const elFormRef = ref()

// 初始化方法
const init = async () => {
 // 建议通过url传参获取目标数据ID 调用 find方法进行查询数据操作 从而决定本页面是create还是update 以下为id作为url参数示例
    if (route.query.id) {
      const res = await findSugarAgents({ ID: route.query.id })
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
               res = await createSugarAgents(formData.value)
               break
             case 'update':
               res = await updateSugarAgents(formData.value)
               break
             default:
               res = await createSugarAgents(formData.value)
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
