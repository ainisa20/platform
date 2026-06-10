<script setup lang="ts">
import { ref, reactive, watch, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import {
  getFinAccountList,
  createFinAccount,
  updateFinAccount,
  deleteFinAccount,
} from '@/api/shop/account'
import type {
  ShopFinAccountResp,
  ShopFinAccountConfig,
  ShopFinAccountCreateReq,
  ShopFinAccountUpdateReq,
} from '@/types/system'

const loading = ref(false)
const tableData = ref<ShopFinAccountResp[]>([])
const total = ref(0)

const searchForm = reactive({
  account_name: '',
  account_type: null as number | null,
  status: null as number | null,
})

const pagination = reactive({ page: 1, page_size: 10 })

async function fetchList() {
  loading.value = true
  try {
    const res = await getFinAccountList({
      page: pagination.page,
      page_size: pagination.page_size,
      account_name: searchForm.account_name || undefined,
      account_type: searchForm.account_type ?? undefined,
      status: searchForm.status ?? undefined,
    })
    tableData.value = res.data.data.list
    total.value = res.data.data.total
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  pagination.page = 1
  fetchList()
}

function handleReset() {
  searchForm.account_name = ''
  searchForm.account_type = null
  searchForm.status = null
  pagination.page = 1
  fetchList()
}

function handlePageChange(page: number) {
  pagination.page = page
  fetchList()
}

function handleSizeChange(size: number) {
  pagination.page_size = size
  pagination.page = 1
  fetchList()
}

const dialogVisible = ref(false)
const dialogTitle = ref('新增账户')
const isEdit = ref(false)
const editId = ref<number>(0)
const submitLoading = ref(false)
const formRef = ref<FormInstance>()

interface AccountFormData {
  account_name: string
  account_type: number
  account_no: string
  initial_balance: number
  status: number
}

const defaultForm = (): AccountFormData => ({
  account_name: '',
  account_type: 1,
  account_no: '',
  initial_balance: 0,
  status: 1,
})

const formData = reactive<AccountFormData>(defaultForm())

const configForm = reactive<ShopFinAccountConfig>({
  bank_name: '',
  branch: '',
  mch_id: '',
  appid: '',
  api_key: '',
  app_id: '',
  merchant_id: '',
  private_key_path: '',
})

function resetConfigForm() {
  Object.assign(configForm, {
    bank_name: '',
    branch: '',
    mch_id: '',
    appid: '',
    api_key: '',
    app_id: '',
    merchant_id: '',
    private_key_path: '',
  })
}

const formRules: FormRules = {
  account_name: [{ required: true, message: '请输入账户名称', trigger: 'blur' }],
  account_type: [{ required: true, message: '请选择账户类型', trigger: 'change' }],
  initial_balance: [{ required: true, message: '请输入初始余额', trigger: 'blur' }],
}

watch(() => formData.account_type, (newVal, oldVal) => {
  if (!isEdit.value && newVal !== oldVal) {
    resetConfigForm()
  }
})

function buildConfigByType(type: number): ShopFinAccountConfig {
  if (type === 1) {
    return {
      bank_name: configForm.bank_name || undefined,
      branch: configForm.branch || undefined,
    }
  }
  if (type === 2) {
    return {
      mch_id: configForm.mch_id || undefined,
      appid: configForm.appid || undefined,
      api_key: configForm.api_key || undefined,
    }
  }
  return {
    app_id: configForm.app_id || undefined,
    merchant_id: configForm.merchant_id || undefined,
    private_key_path: configForm.private_key_path || undefined,
  }
}

function openCreateDialog() {
  isEdit.value = false
  dialogTitle.value = '新增账户'
  editId.value = 0
  Object.assign(formData, defaultForm())
  resetConfigForm()
  dialogVisible.value = true
}

function openEditDialog(row: ShopFinAccountResp) {
  isEdit.value = true
  dialogTitle.value = '编辑账户'
  editId.value = row.id
  Object.assign(formData, {
    account_name: row.account_name,
    account_type: row.account_type,
    account_no: row.account_no,
    initial_balance: row.initial_balance,
    status: row.status,
  })
  resetConfigForm()
  Object.assign(configForm, row.config || {})
  dialogVisible.value = true
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    if (isEdit.value) {
      const updateData: ShopFinAccountUpdateReq = {
        account_name: formData.account_name,
        account_no: formData.account_no || undefined,
        config: buildConfigByType(formData.account_type),
        status: formData.status,
      }
      await updateFinAccount(editId.value, updateData)
      ElMessage.success('更新成功')
    } else {
      const createData: ShopFinAccountCreateReq = {
        account_name: formData.account_name,
        account_type: formData.account_type,
        account_no: formData.account_no || undefined,
        initial_balance: formData.initial_balance,
        config: buildConfigByType(formData.account_type),
        status: formData.status,
      }
      await createFinAccount(createData)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchList()
  } finally {
    submitLoading.value = false
  }
}

async function handleDelete(row: ShopFinAccountResp) {
  await ElMessageBox.confirm(
    `确定要删除账户「${row.account_name}」吗？`,
    '删除确认',
    { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
  )
  await deleteFinAccount(row.id)
  ElMessage.success('删除成功')
  fetchList()
}

function formatTime(val: string | null): string {
  if (!val) return '-'
  return val.replace('T', ' ').substring(0, 19)
}

function formatAmount(val: number | null | undefined): string {
  if (val == null) return '-'
  return '￥' + Number(val).toFixed(2)
}

const accountTypeTagType = (t: number) => {
  if (t === 1) return 'primary'
  if (t === 2) return 'success'
  return 'warning'
}
const accountTypeText = (t: number) => {
  if (t === 1) return '对公'
  if (t === 2) return '微信'
  return '支付宝'
}
const statusTagType = (s: number) => (s === 1 ? 'success' : 'info')
const statusText = (s: number) => (s === 1 ? '启用' : '停用')

onMounted(fetchList)
</script>

<template>
  <div class="account-page">
    <el-card shadow="never" class="search-card">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="账户名称">
          <el-input v-model="searchForm.account_name" placeholder="请输入账户名称" clearable style="width: 200px" />
        </el-form-item>
        <el-form-item label="账户类型">
          <el-select v-model="searchForm.account_type" placeholder="全部类型" clearable style="width: 140px">
            <el-option label="对公" :value="1" />
            <el-option label="微信" :value="2" />
            <el-option label="支付宝" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部状态" clearable style="width: 120px">
            <el-option label="启用" :value="1" />
            <el-option label="停用" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never" class="table-card">
      <div class="table-header">
        <el-button v-permission="'shop:finance:account:create'" type="primary" @click="openCreateDialog">
          新增账户
        </el-button>
      </div>

      <el-table v-loading="loading" :data="tableData" border stripe style="width: 100%">
        <el-table-column prop="id" label="序号" width="70" />
        <el-table-column prop="account_name" label="账户名称" min-width="150" />
        <el-table-column label="账户类型" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="accountTypeTagType(row.account_type)" size="small">
              {{ accountTypeText(row.account_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="账号" min-width="150">
          <template #default="{ row }">
            {{ row.account_no || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="初始余额" width="130" align="right">
          <template #default="{ row }">
            {{ formatAmount(row.initial_balance) }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">
              {{ statusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建人" min-width="100">
          <template #default="{ row }">
            {{ row.created_by_name || row.created_by || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="创建时间" min-width="170">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button v-permission="'shop:finance:account:update'" type="primary" link size="small" @click="openEditDialog(row)">
              编辑
            </el-button>
            <el-button v-permission="'shop:finance:account:delete'" type="danger" link size="small" @click="handleDelete(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.page_size"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          background
          @current-change="handlePageChange"
          @size-change="handleSizeChange"
        />
      </div>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="640px" :close-on-click-modal="false" destroy-on-close>
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="100px">
        <el-form-item label="账户名称" prop="account_name">
          <el-input v-model="formData.account_name" placeholder="请输入账户名称" />
        </el-form-item>

        <el-form-item v-if="!isEdit" label="账户类型" prop="account_type">
          <el-radio-group v-model="formData.account_type">
            <el-radio :value="1">对公</el-radio>
            <el-radio :value="2">微信</el-radio>
            <el-radio :value="3">支付宝</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="账号">
          <el-input v-model="formData.account_no" placeholder="请输入账号" />
        </el-form-item>

        <el-form-item v-if="!isEdit" label="初始余额" prop="initial_balance">
          <el-input-number v-model="formData.initial_balance" :precision="2" :min="0" :step="100" style="width: 100%" />
        </el-form-item>

        <div class="config-section">
          <div class="config-title">账户配置</div>
          <template v-if="formData.account_type === 1">
            <el-form-item label="开户行">
              <el-input v-model="configForm.bank_name" placeholder="请输入开户行" />
            </el-form-item>
            <el-form-item label="支行">
              <el-input v-model="configForm.branch" placeholder="请输入支行" />
            </el-form-item>
          </template>
          <template v-else-if="formData.account_type === 2">
            <el-form-item label="商户号">
              <el-input v-model="configForm.mch_id" placeholder="请输入商户号" />
            </el-form-item>
            <el-form-item label="AppID">
              <el-input v-model="configForm.appid" placeholder="请输入 AppID" />
            </el-form-item>
            <el-form-item label="API Key">
              <el-input v-model="configForm.api_key" placeholder="请输入 API Key" show-password />
            </el-form-item>
          </template>
          <template v-else>
            <el-form-item label="AppID">
              <el-input v-model="configForm.app_id" placeholder="请输入 AppID" />
            </el-form-item>
            <el-form-item label="商户号">
              <el-input v-model="configForm.merchant_id" placeholder="请输入商户号" />
            </el-form-item>
            <el-form-item label="私钥路径">
              <el-input v-model="configForm.private_key_path" placeholder="请输入私钥文件路径" />
            </el-form-item>
          </template>
        </div>

        <el-form-item label="状态">
          <el-radio-group v-model="formData.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="2">停用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.account-page { padding: 16px; }
.search-card { margin-bottom: 16px; }
.search-card :deep(.el-card__body) { padding-bottom: 2px; }
.table-header { margin-bottom: 16px; }
.pagination-wrap { display: flex; justify-content: flex-end; margin-top: 16px; }
.config-section {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  padding: 12px 16px 4px;
  margin-bottom: 18px;
  background: var(--el-fill-color-blank);
}
.config-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  margin-bottom: 8px;
  letter-spacing: 0.3px;
  border-left: 3px solid var(--el-color-primary);
  padding-left: 8px;
}
</style>
