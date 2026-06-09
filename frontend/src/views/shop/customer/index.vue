<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import {
  getCustomerList,
  createCustomer,
  updateCustomer,
  deleteCustomer,
  exportCustomers,
  getCustomerOrders,
} from '@/api/shop/customer'
import type {
  ShopCustomerResp,
  ShopCustomerCreateReq,
  ShopCustomerUpdateReq,
  ShopCustomerOrderResp,
} from '@/types/system'

const loading = ref(false)
const tableData = ref<ShopCustomerResp[]>([])
const total = ref(0)

const searchForm = reactive({
  customer_name: '',
  contact_person: '',
  status: null as number | null,
})

const pagination = reactive({ page: 1, page_size: 10 })

async function fetchList() {
  loading.value = true
  try {
    const res = await getCustomerList({
      page: pagination.page,
      page_size: pagination.page_size,
      customer_name: searchForm.customer_name || undefined,
      contact_person: searchForm.contact_person || undefined,
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
  searchForm.customer_name = ''
  searchForm.contact_person = ''
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
const dialogTitle = ref('新增客户')
const isEdit = ref(false)
const editId = ref<number>(0)
const submitLoading = ref(false)
const formRef = ref<FormInstance>()

interface CustomerFormData {
  customer_name: string
  customer_type: number
  contact_person: string
  contact_phone: string
  contact_email: string
  address: string
  remark: string
  status: number
}

const defaultForm = (): CustomerFormData => ({
  customer_name: '',
  customer_type: 1,
  contact_person: '',
  contact_phone: '',
  contact_email: '',
  address: '',
  remark: '',
  status: 1,
})

const formData = reactive<CustomerFormData>(defaultForm())

const formRules = reactive<FormRules>({
  customer_name: [{ required: true, message: '请输入客户名称', trigger: 'blur' }],
  customer_type: [{ required: true, message: '请选择客户类型', trigger: 'change' }],
  contact_email: [{ type: 'email', message: '请输入正确的邮箱地址', trigger: 'blur' }],
})

function openCreateDialog() {
  isEdit.value = false
  dialogTitle.value = '新增客户'
  editId.value = 0
  Object.assign(formData, defaultForm())
  dialogVisible.value = true
}

function openEditDialog(row: ShopCustomerResp) {
  isEdit.value = true
  dialogTitle.value = '编辑客户'
  editId.value = row.id
  Object.assign(formData, {
    customer_name: row.customer_name,
    customer_type: row.customer_type,
    contact_person: row.contact_person,
    contact_phone: row.contact_phone,
    contact_email: row.contact_email,
    address: row.address,
    remark: row.remark,
    status: row.status,
  })
  dialogVisible.value = true
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    if (isEdit.value) {
      const updateData: ShopCustomerUpdateReq = {
        customer_name: formData.customer_name,
        customer_type: formData.customer_type,
        contact_person: formData.contact_person || undefined,
        contact_phone: formData.contact_phone || undefined,
        contact_email: formData.contact_email || undefined,
        address: formData.address || undefined,
        remark: formData.remark || undefined,
        status: formData.status,
      }
      await updateCustomer(editId.value, updateData)
      ElMessage.success('更新成功')
    } else {
      const createData: ShopCustomerCreateReq = {
        customer_name: formData.customer_name,
        customer_type: formData.customer_type,
        contact_person: formData.contact_person || undefined,
        contact_phone: formData.contact_phone || undefined,
        contact_email: formData.contact_email || undefined,
        address: formData.address || undefined,
        remark: formData.remark || undefined,
        status: formData.status,
      }
      await createCustomer(createData)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchList()
  } finally {
    submitLoading.value = false
  }
}

async function handleDelete(row: ShopCustomerResp) {
  await ElMessageBox.confirm(
    `确定要删除客户「${row.customer_name}」吗？`,
    '删除确认',
    { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
  )
  await deleteCustomer(row.id)
  ElMessage.success('删除成功')
  fetchList()
}

const orderDialogVisible = ref(false)
const orderLoading = ref(false)
const orderCustomerId = ref(0)
const orderCustomerName = ref('')
const orderList = ref<ShopCustomerOrderResp[]>([])

async function openOrderDialog(row: ShopCustomerResp) {
  orderCustomerId.value = row.id
  orderCustomerName.value = row.customer_name
  orderList.value = []
  orderDialogVisible.value = true
  orderLoading.value = true
  try {
    const res = await getCustomerOrders(row.id)
    orderList.value = res.data.data || []
  } finally {
    orderLoading.value = false
  }
}

async function handleExport() {
  const res = await exportCustomers({
    customer_name: searchForm.customer_name || undefined,
    contact_person: searchForm.contact_person || undefined,
    status: searchForm.status ?? undefined,
  })
  const data = res.data.data
  if (!data || data.length === 0) {
    ElMessage.warning('没有数据可导出')
    return
  }
  const headers = ['客户名称', '客户类型', '联系人', '电话', '邮箱', '地址', '状态', '创建人', '创建时间']
  const rows = data.map(c => [
    c.customer_name,
    c.customer_type === 1 ? '个人' : '企业',
    c.contact_person,
    c.contact_phone,
    c.contact_email,
    c.address,
    c.status === 1 ? '启用' : '停用',
    String(c.created_by),
    formatTime(c.created_at),
  ])
  const csv = [headers, ...rows].map(r => r.map(cell => `"${String(cell).replace(/"/g, '""')}"`).join(',')).join('\n')
  const blob = new Blob(['\uFEFF' + csv], { type: 'text/csv;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `客户列表_${new Date().toISOString().slice(0, 10)}.csv`
  link.click()
  URL.revokeObjectURL(url)
  ElMessage.success('导出成功')
}

function formatTime(val: string | null): string {
  if (!val) return '-'
  return val.replace('T', ' ').substring(0, 19)
}

function formatAmount(val: number): string {
  return Number(val).toFixed(2)
}

const customerTypeTagType = (type: number) => (type === 1 ? 'info' : 'warning')
const customerTypeText = (type: number) => (type === 1 ? '个人' : '企业')

const statusTagType = (status: number) => (status === 1 ? 'success' : 'info')
const statusText = (status: number) => (status === 1 ? '启用' : '停用')

const orderStatusTagType = (status: number) => {
  if (status === 1) return 'info'
  if (status === 2) return 'warning'
  if (status === 3) return 'success'
  return 'danger'
}
const orderStatusText = (status: number) => {
  if (status === 1) return '待支付'
  if (status === 2) return '进行中'
  if (status === 3) return '已完成'
  return '已取消'
}

onMounted(() => {
  fetchList()
})
</script>

<template>
  <div class="customer-page">
    <el-card shadow="never" class="search-card">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="客户名称">
          <el-input v-model="searchForm.customer_name" placeholder="请输入客户名称" clearable />
        </el-form-item>
        <el-form-item label="联系人">
          <el-input v-model="searchForm.contact_person" placeholder="请输入联系人" clearable />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部" clearable style="width: 120px">
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
        <el-button v-permission="'shop:customer:create'" type="primary" @click="openCreateDialog">
          新增客户
        </el-button>
        <el-button v-permission="'shop:customer:export'" type="success" @click="handleExport">
          导出
        </el-button>
      </div>

      <el-table v-loading="loading" :data="tableData" border stripe style="width: 100%">
        <el-table-column prop="id" label="序号" width="70" align="center" />
        <el-table-column prop="customer_name" label="客户名称" min-width="150" />
        <el-table-column label="客户类型" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="customerTypeTagType(row.customer_type)" size="small">
              {{ customerTypeText(row.customer_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="contact_person" label="联系人" min-width="100">
          <template #default="{ row }">
            {{ row.contact_person || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="contact_phone" label="联系电话" min-width="130">
          <template #default="{ row }">
            {{ row.contact_phone || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="contact_email" label="邮箱" min-width="160">
          <template #default="{ row }">
            {{ row.contact_email || '-' }}
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
            {{ row.created_by || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="创建时间" min-width="170">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button v-permission="'shop:customer:update'" type="primary" link size="small" @click="openEditDialog(row)">
              编辑
            </el-button>
            <el-button v-permission="'shop:customer:list'" type="warning" link size="small" @click="openOrderDialog(row)">
              订单
            </el-button>
            <el-button v-permission="'shop:customer:delete'" type="danger" link size="small" @click="handleDelete(row)">
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

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="560px" :close-on-click-modal="false" destroy-on-close>
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="90px">
        <el-form-item label="客户名称" prop="customer_name">
          <el-input v-model="formData.customer_name" placeholder="请输入客户名称" />
        </el-form-item>
        <el-form-item label="客户类型" prop="customer_type">
          <el-radio-group v-model="formData.customer_type">
            <el-radio :value="1">个人</el-radio>
            <el-radio :value="2">企业</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="联系人" prop="contact_person">
          <el-input v-model="formData.contact_person" placeholder="请输入联系人" />
        </el-form-item>
        <el-form-item label="联系电话" prop="contact_phone">
          <el-input v-model="formData.contact_phone" placeholder="请输入联系电话" />
        </el-form-item>
        <el-form-item label="邮箱" prop="contact_email">
          <el-input v-model="formData.contact_email" type="email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="地址" prop="address">
          <el-input v-model="formData.address" placeholder="请输入地址" />
        </el-form-item>
        <el-form-item label="备注" prop="remark">
          <el-input v-model="formData.remark" type="textarea" :rows="3" placeholder="请输入备注" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
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

    <el-dialog
      v-model="orderDialogVisible"
      :title="`客户订单 - ${orderCustomerName}`"
      width="800px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <div v-loading="orderLoading">
        <el-table v-if="orderList.length > 0" :data="orderList" border stripe style="width: 100%">
          <el-table-column prop="order_no" label="订单号" min-width="180" />
          <el-table-column label="总金额" min-width="120" align="right">
            <template #default="{ row }">
              {{ formatAmount(row.total_amount) }}
            </template>
          </el-table-column>
          <el-table-column label="状态" width="100" align="center">
            <template #default="{ row }">
              <el-tag :type="orderStatusTagType(row.status)" size="small">
                {{ orderStatusText(row.status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="创建人" min-width="100">
            <template #default="{ row }">
              {{ row.created_by || '-' }}
            </template>
          </el-table-column>
          <el-table-column label="创建时间" min-width="170">
            <template #default="{ row }">
              {{ formatTime(row.created_at) }}
            </template>
          </el-table-column>
        </el-table>
        <el-empty v-else description="暂无订单" />
      </div>
      <template #footer>
        <el-button @click="orderDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.customer-page { padding: 16px; }
.search-card { margin-bottom: 16px; }
.search-card :deep(.el-card__body) { padding-bottom: 2px; }
.table-header { margin-bottom: 16px; display: flex; gap: 8px; }
.pagination-wrap { display: flex; justify-content: flex-end; margin-top: 16px; }
</style>
