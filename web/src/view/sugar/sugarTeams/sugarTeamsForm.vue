
<template>
  <div>
    <div class="gva-form-box">
      <el-form :model="formData" ref="elFormRef" label-position="right" :rules="rule" label-width="80px">
        <el-form-item label="id字段:" prop="id">
    <el-input v-model="formData.id" :clearable="true" placeholder="请输入id字段" />
</el-form-item>
        <el-form-item label="teamName字段:" prop="teamName">
    <el-input v-model="formData.teamName" :clearable="true" placeholder="请输入teamName字段" />
</el-form-item>
        <el-form-item label="团队创建者/个人空间的所有者:" prop="ownerId">
    <el-input v-model="formData.ownerId" :clearable="true" placeholder="请输入团队创建者/个人空间的所有者" />
</el-form-item>
        <el-form-item label="是否为个人空间团队 (true代表个人空间):" prop="isPersonal">
    <el-switch v-model="formData.isPersonal" active-color="#13ce66" inactive-color="#ff4949" active-text="是" inactive-text="否" clearable ></el-switch>
</el-form-item>
        <el-form-item label="createdBy字段:" prop="createdBy">
    <el-input v-model="formData.createdBy" :clearable="true" placeholder="请输入createdBy字段" />
</el-form-item>
        <el-form-item label="createdAt字段:" prop="createdAt">
    <el-date-picker v-model="formData.createdAt" type="date" style="width:100%" placeholder="选择日期" :clearable="true" />
</el-form-item>
        <el-form-item label="updatedBy字段:" prop="updatedBy">
    <el-input v-model="formData.updatedBy" :clearable="true" placeholder="请输入updatedBy字段" />
</el-form-item>
        <el-form-item label="updatedAt字段:" prop="updatedAt">
    <el-date-picker v-model="formData.updatedAt" type="date" style="width:100%" placeholder="选择日期" :clearable="true" />
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
  createSugarTeams,
  updateSugarTeams,
  findSugarTeams
} from '@/api/sugar/sugarTeams'

defineOptions({
    name: 'SugarTeamsForm'
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
            id: '',
            teamName: '',
            ownerId: '',
            isPersonal: false,
            createdBy: '',
            createdAt: new Date(),
            updatedBy: '',
            updatedAt: new Date(),
        })
// 验证规则
const rule = reactive({
})

const elFormRef = ref()

// 初始化方法
const init = async () => {
 // 建议通过url传参获取目标数据ID 调用 find方法进行查询数据操作 从而决定本页面是create还是update 以下为id作为url参数示例
    if (route.query.id) {
      const res = await findSugarTeams({ ID: route.query.id })
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
               res = await createSugarTeams(formData.value)
               break
             case 'update':
               res = await updateSugarTeams(formData.value)
               break
             default:
               res = await createSugarTeams(formData.value)
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
