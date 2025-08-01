
<template>
  <div>
    <div class="gva-form-box">
      <el-form :model="formData" ref="elFormRef" label-position="right" :rules="rule" label-width="80px">
        <el-form-item label="模型的业务名称, 如“季度销售报告”:" prop="name">
    <el-input v-model="formData.name" :clearable="true" placeholder="请输入模型的业务名称, 如“季度销售报告”" />
</el-form-item>
        <el-form-item label="description字段:" prop="description">
    <el-input v-model="formData.description" :clearable="true" placeholder="请输入description字段" />
</el-form-item>
        <el-form-item label="teamId字段:" prop="teamId">
    <el-input v-model="formData.teamId" :clearable="true" placeholder="请输入teamId字段" />
</el-form-item>
        <el-form-item label="关联的数据库连接:" prop="connectionId">
    <el-input v-model="formData.connectionId" :clearable="true" placeholder="请输入关联的数据库连接" />
</el-form-item>
        <el-form-item label="源数据库中的真实表名:" prop="sourceTableName">
    <el-input v-model="formData.sourceTableName" :clearable="true" placeholder="请输入源数据库中的真实表名" />
</el-form-item>
        <el-form-item label="查询参数配置, 定义用户可用的筛选条件:" prop="parameterConfig">
    // 此字段为json结构，可以前端自行控制展示和数据绑定模式 需绑定json的key为 formData.parameterConfig 后端会按照json的类型进行存取
    {{ formData.parameterConfig }}
</el-form-item>
        <el-form-item label="可返回字段配置, 定义用户可获取的数据列:" prop="returnableColumnsConfig">
    // 此字段为json结构，可以前端自行控制展示和数据绑定模式 需绑定json的key为 formData.returnableColumnsConfig 后端会按照json的类型进行存取
    {{ formData.returnableColumnsConfig }}
</el-form-item>
        <el-form-item label="用于行级权限判断的字段名, 如 city_code:" prop="permissionKeyColumn">
    <el-input v-model="formData.permissionKeyColumn" :clearable="true" placeholder="请输入用于行级权限判断的字段名, 如 city_code" />
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
  createSugarSemanticModels,
  updateSugarSemanticModels,
  findSugarSemanticModels
} from '@/api/sugar/sugarSemanticModels'

defineOptions({
    name: 'SugarSemanticModelsForm'
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
            teamId: '',
            connectionId: '',
            sourceTableName: '',
            parameterConfig: {},
            returnableColumnsConfig: {},
            permissionKeyColumn: '',
        })
// 验证规则
const rule = reactive({
})

const elFormRef = ref()

// 初始化方法
const init = async () => {
 // 建议通过url传参获取目标数据ID 调用 find方法进行查询数据操作 从而决定本页面是create还是update 以下为id作为url参数示例
    if (route.query.id) {
      const res = await findSugarSemanticModels({ ID: route.query.id })
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
               res = await createSugarSemanticModels(formData.value)
               break
             case 'update':
               res = await updateSugarSemanticModels(formData.value)
               break
             default:
               res = await createSugarSemanticModels(formData.value)
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
