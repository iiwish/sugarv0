
<template>
  <div>
    <div class="gva-form-box">
      <el-form :model="formData" ref="elFormRef" label-position="right" :rules="rule" label-width="80px">
        <el-form-item label="name字段:" prop="name">
    <el-input v-model="formData.name" :clearable="true" placeholder="请输入name字段" />
</el-form-item>
        <el-form-item label="teamId字段:" prop="teamId">
    <el-input v-model="formData.teamId" :clearable="true" placeholder="请输入teamId字段" />
</el-form-item>
        <el-form-item label="dbType字段:" prop="dbType">
    <el-input v-model="formData.dbType" :clearable="true" placeholder="请输入dbType字段" />
</el-form-item>
        <el-form-item label="是否为Sugar内部数据库 (true代表本库):" prop="isInternal">
    <el-switch v-model="formData.isInternal" active-color="#13ce66" inactive-color="#ff4949" active-text="是" inactive-text="否" clearable ></el-switch>
</el-form-item>
        <el-form-item label="内部数据库可为null:" prop="host">
    <el-input v-model="formData.host" :clearable="true" placeholder="请输入内部数据库可为null" />
</el-form-item>
        <el-form-item label="内部数据库可为null:" prop="port">
    <el-input v-model.number="formData.port" :clearable="true" placeholder="请输入内部数据库可为null" />
</el-form-item>
        <el-form-item label="内部数据库可为null:" prop="username">
    <el-input v-model="formData.username" :clearable="true" placeholder="请输入内部数据库可为null" />
</el-form-item>
        <el-form-item label="内部数据库可为null:" prop="encryptedPassword">
    <el-input v-model="formData.encryptedPassword" :clearable="true" placeholder="请输入内部数据库可为null" />
</el-form-item>
        <el-form-item label="databaseName字段:" prop="databaseName">
    <el-input v-model="formData.databaseName" :clearable="true" placeholder="请输入databaseName字段" />
</el-form-item>
        <el-form-item label="sslConfig字段:" prop="sslConfig">
    <el-select v-model="formData.sslConfig" placeholder="请选择sslConfig字段" style="width:100%" filterable :clearable="true">
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
  createSugarDbConnections,
  updateSugarDbConnections,
  findSugarDbConnections
} from '@/api/sugar/sugarDbConnections'

defineOptions({
    name: 'SugarDbConnectionsForm'
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
            teamId: '',
            dbType: '',
            isInternal: false,
            host: '',
            port: undefined,
            username: '',
            encryptedPassword: '',
            databaseName: '',
            sslConfig: null,
        })
// 验证规则
const rule = reactive({
})

const elFormRef = ref()

// 初始化方法
const init = async () => {
 // 建议通过url传参获取目标数据ID 调用 find方法进行查询数据操作 从而决定本页面是create还是update 以下为id作为url参数示例
    if (route.query.id) {
      const res = await findSugarDbConnections({ ID: route.query.id })
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
               res = await createSugarDbConnections(formData.value)
               break
             case 'update':
               res = await updateSugarDbConnections(formData.value)
               break
             default:
               res = await createSugarDbConnections(formData.value)
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
