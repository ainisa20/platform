<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Document } from '@element-plus/icons-vue'
import type { FormInstance, FormRules } from 'element-plus'
import {
  getRecordList,
  getRecord,
  createRecord,
  updateRecord,
  deleteRecord,
  reviewRecord,
  exportRecords,
  getRecordAttachments,
  createRecordAttachment,
  downloadRecordAttachment,
} from '@/api/shop/record'
import { getFinAccountList } from '@/api/shop/account'
import { getShopFinCategories } from '@/api/shop/finance'
import { getOrderList } from '@/api/shop/order'
import type {
  FinanceRecordResp,
  FinanceRecordCreateReq,
  FinanceRecordUpdateReq,
  FinanceReviewReq,
  FinanceAttachmentResp,
  ShopFinAccountResp,
  ShopFinCategoryResp,
  OrderResp,
} from '@/types/system'

const loading = ref(false)
const tableData = ref<FinanceRecordResp[]>([])
const total = ref(0)

const flatCategories = ref<ShopFinCategoryResp[]>([])

function flattenTree(nodes: ShopFinCategoryResp[]): ShopFinCategoryResp[] {
  const flat: ShopFinCategoryResp[] = []
  for (const n of nodes) {
    const { children, ...rest } = n
    flat.push(rest)
    if (children?.length) flat.push(...flattenTree(children))
  }
  return flat
}

function findCategoryPath(leafId: number) {
  const leaf = flatCategories.value.find(c => c.id === leafId)
  if (!leaf) return null
  const l2 = flatCategories.value.find(c => c.id === leaf.parent_id)
  const l1 = l2 ? flatCategories.value.find(c => c.id === l2.parent_id) : null
  return {
    l1_id: l1?.id ?? null,
    l2_id: l2?.id ?? null,
    l3_id: leaf.id,
    record_type: leaf.category_type,
  }
}

const searchForm = reactive({
  record_no: '',
  category_l1: '',
  category_l2: '',
  category_l3: '',
  review_status: null as number | null,
})

// Search category cascade options
const searchL1Options = computed(() => flatCategories.value.filter(c => c.level === 1))
const searchL2Options = computed(() => {
  if (!searchForm.category_l1) return []
  const l1 = flatCategories.value.find(c => c.level === 1 && c.category_name === searchForm.category_l1)
  if (!l1) return []
  return flatCategories.value.filter(c => c.level === 2 && c.parent_id === l1.id)
})
const searchL3Options = computed(() => {
  if (!searchForm.category_l2) return []
  const l2 = flatCategories.value.find(c => c.level === 2 && c.category_name === searchForm.category_l2)
  if (!l2) return []
  return flatCategories.value.filter(c => c.level === 3 && c.parent_id === l2.id)
})

const pagination = reactive({ page: 1, page_size: 10 })

async function fetchList() {
  loading.value = true
  try {
    const res = await getRecordList({
      page: pagination.page,
      page_size: pagination.page_size,
      record_no: searchForm.record_no || undefined,
      category_l1: searchForm.category_l1 || undefined,
      category_l2: searchForm.category_l2 || undefined,
      category_l3: searchForm.category_l3 || undefined,
      review_status: searchForm.review_status ?? undefined,
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
  searchForm.record_no = ''
  searchForm.category_l1 = ''
  searchForm.category_l2 = ''
  searchForm.category_l3 = ''
  searchForm.review_status = null
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

const reviewStatusTagType = (s: number) => {
  if (s === 1) return 'info'
  if (s === 2) return 'success'
  return 'danger'
}
const reviewStatusText = (s: number) => {
  if (s === 1) return '待审核'
  if (s === 2) return '已通过'
  return '已驳回'
}
const recordTypeTagType = (t: number) => (t === 1 ? 'success' : 'warning')
const recordTypeText = (t: number) => (t === 1 ? '收入' : '支出')

function formatTime(val: string | null): string {
  if (!val) return '-'
  return val.replace('T', ' ').substring(0, 19)
}
function formatAmount(val: number | null | undefined): string {
  if (val == null) return '-'
  return '￥' + Number(val).toFixed(2)
}
function formatFileSize(bytes: number): string {
  if (!bytes) return '-'
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(2) + ' KB'
  return (bytes / 1024 / 1024).toFixed(2) + ' MB'
}

async function handleExport() {
  const res = await exportRecords({
    record_no: searchForm.record_no || undefined,
    review_status: searchForm.review_status ?? undefined,
  })
  const data = res.data.data
  if (!data || data.length === 0) {
    ElMessage.warning('没有数据可导出')
    return
  }
  const headers = ['记录号', '账户', '一级分类', '二级分类', '三级分类', '类型', '金额', '实际金额', '审核状态', '记账日期', '创建人', '创建时间']
  const rows = data.map((r) => [
    r.record_no,
    r.account_name,
    r.category_l1,
    r.category_l2,
    r.category_l3,
    recordTypeText(r.record_type),
    formatAmount(r.amount),
    formatAmount(r.actual_amount),
    reviewStatusText(r.review_status),
    r.record_date,
    r.created_by_name || String(r.created_by),
    formatTime(r.created_at),
  ])
  const csv = [headers, ...rows]
    .map((r) => r.map((cell) => `"${String(cell).replace(/"/g, '""')}"`).join(','))
    .join('\n')
  const blob = new Blob(['\uFEFF' + csv], { type: 'text/csv;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `收支记录_${new Date().toISOString().slice(0, 10)}.csv`
  link.click()
  URL.revokeObjectURL(url)
  ElMessage.success('导出成功')
}

const dialogVisible = ref(false)
const dialogTitle = ref('新建记录')
const isEdit = ref(false)
const editId = ref<number>(0)
const submitLoading = ref(false)
const formRef = ref<FormInstance>()

const accountOptions = ref<ShopFinAccountResp[]>([])
const categoryOptions = ref<ShopFinCategoryResp[]>([])
const orderOptions = ref<OrderResp[]>([])
const isReadonly = ref(false)

interface RecordFormData {
  account_id: number | null
  category_id: number | null
  category_l1_id: number | null
  category_l2_id: number | null
  record_type: number
  amount: number
  order_group_id: number | null
  record_date: string
  remark: string
}

const defaultForm = (): RecordFormData => ({
  account_id: null,
  category_id: null,
  category_l1_id: null,
  category_l2_id: null,
  record_type: 1,
  amount: 0,
  order_group_id: null,
  record_date: new Date().toISOString().slice(0, 10),
  remark: '',
})

const formData = reactive<RecordFormData>(defaultForm())

const formRules: FormRules = {
  account_id: [{ required: true, message: '请选择账户', trigger: 'change' }],
  category_id: [{ required: true, message: '请选择三级分类', trigger: 'change' }],
  amount: [
    {
      required: true,
      validator: (_rule: unknown, value: number, callback: (err?: Error) => void) => {
        if (value == null || Number(value) <= 0) {
          callback(new Error('请输入金额'))
        } else {
          callback()
        }
      },
      trigger: 'blur',
    },
  ],
  record_date: [{ required: true, message: '请选择记账日期', trigger: 'change' }],
}

async function loadAccounts() {
  const res = await getFinAccountList({ page: 1, page_size: 1000, status: 1 })
  accountOptions.value = res.data.data.list
}

async function loadCategories() {
  const res = await getShopFinCategories()
  categoryOptions.value = res.data.data
  flatCategories.value = flattenTree(res.data.data)
}

async function loadOrders() {
  const res = await getOrderList({ page: 1, page_size: 1000 })
  orderOptions.value = res.data.data.list
}

// Dialog category cascade options
const dialogL1Options = computed(() => flatCategories.value.filter(c => c.level === 1))
const dialogL2Options = computed(() => {
  if (!formData.category_l1_id) return []
  return flatCategories.value.filter(c => c.level === 2 && c.parent_id === formData.category_l1_id)
})
const dialogL3Options = computed(() => {
  if (!formData.category_l2_id) return []
  return flatCategories.value.filter(c => c.level === 3 && c.parent_id === formData.category_l2_id)
})

function onSearchL1Change() {
  searchForm.category_l2 = ''
  searchForm.category_l3 = ''
}
function onSearchL2Change() {
  searchForm.category_l3 = ''
}

function onDialogL1Change() {
  formData.category_l2_id = null
  formData.category_id = null
  formData.record_type = 1
}
function onDialogL2Change() {
  formData.category_id = null
  formData.record_type = 1
}
function onDialogL3Change(val: number) {
  const cat = flatCategories.value.find(c => c.id === val)
  if (cat) {
    formData.record_type = cat.category_type
  }
}

function resetForm() {
  Object.assign(formData, defaultForm())
}

async function openCreateDialog() {
  isEdit.value = false
  isReadonly.value = false
  dialogTitle.value = '新建记录'
  editId.value = 0
  resetForm()
  dialogVisible.value = true
  await Promise.all([loadAccounts(), loadCategories(), loadOrders()])
}

async function openEditDialog(row: FinanceRecordResp) {
  if (row.review_status === 2) {
    ElMessage.warning('已审核通过的记录不可编辑')
    return
  }
  isEdit.value = true
  isReadonly.value = false
  dialogTitle.value = '编辑记录'
  editId.value = row.id
  await Promise.all([loadAccounts(), loadCategories(), loadOrders()])
  const res = await getRecord(row.id)
  const r = res.data.data
  const path = findCategoryPath(r.category_id)
  Object.assign(formData, {
    account_id: r.account_id,
    category_l1_id: path?.l1_id ?? null,
    category_l2_id: path?.l2_id ?? null,
    category_id: r.category_id,
    record_type: r.record_type,
    amount: r.amount,
    order_group_id: r.order_group_id,
    record_date: r.record_date ? r.record_date.slice(0, 10) : new Date().toISOString().slice(0, 10),
    remark: r.remark || '',
  })
  dialogVisible.value = true
}

async function openDetailDialog(row: FinanceRecordResp) {
  isEdit.value = false
  isReadonly.value = true
  dialogTitle.value = '记录详情'
  editId.value = row.id
  await Promise.all([loadAccounts(), loadCategories(), loadOrders()])
  const res = await getRecord(row.id)
  const r = res.data.data
  const path = findCategoryPath(r.category_id)
  Object.assign(formData, {
    account_id: r.account_id,
    category_l1_id: path?.l1_id ?? null,
    category_l2_id: path?.l2_id ?? null,
    category_id: r.category_id,
    record_type: r.record_type,
    amount: r.amount,
    order_group_id: r.order_group_id,
    record_date: r.record_date ? r.record_date.slice(0, 10) : new Date().toISOString().slice(0, 10),
    remark: r.remark || '',
  })
  dialogVisible.value = true
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    if (isEdit.value) {
      const data: FinanceRecordUpdateReq = {
        account_id: formData.account_id as number,
        category_id: formData.category_id as number,
        record_type: formData.record_type,
        amount: formData.amount,
        order_group_id: formData.order_group_id ?? null,
        record_date: formData.record_date,
        remark: formData.remark || undefined,
      }
      await updateRecord(editId.value, data)
      ElMessage.success('更新成功')
    } else {
      const data: FinanceRecordCreateReq = {
        account_id: formData.account_id as number,
        category_id: formData.category_id as number,
        record_type: formData.record_type,
        amount: formData.amount,
        order_group_id: formData.order_group_id ?? null,
        record_date: formData.record_date,
        remark: formData.remark || undefined,
      }
      await createRecord(data)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchList()
  } finally {
    submitLoading.value = false
  }
}

async function handleDelete(row: FinanceRecordResp) {
  if (row.review_status === 2) {
    ElMessage.warning('已审核通过的记录不可删除')
    return
  }
  await ElMessageBox.confirm(
    `确定要删除记录「${row.record_no}」吗？`,
    '删除确认',
    { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
  )
  await deleteRecord(row.id)
  ElMessage.success('删除成功')
  fetchList()
}

const reviewDialogVisible = ref(false)
const reviewLoading = ref(false)
const reviewFormRef = ref<FormInstance>()
const currentRecord = ref<FinanceRecordResp | null>(null)
const reviewForm = reactive({
  action: 'approve' as 'approve' | 'reject',
  actual_amount: 0,
  notes: '',
})

const reviewRules = reactive<FormRules>({
  action: [{ required: true, message: '请选择审核动作', trigger: 'change' }],
  actual_amount: [
    {
      required: true,
      validator: (_rule: unknown, value: number, callback: (err?: Error) => void) => {
        if (reviewForm.action === 'approve' && (value == null || Number(value) <= 0)) {
          callback(new Error('请输入实际金额'))
        } else {
          callback()
        }
      },
      trigger: 'blur',
    },
  ],
})

function resetReviewForm() {
  reviewForm.action = 'approve'
  reviewForm.actual_amount = 0
  reviewForm.notes = ''
}

async function openReviewDialog(row: FinanceRecordResp) {
  if (row.review_status === 2) {
    ElMessage.warning('已审核通过的记录无需再次审核')
    return
  }
  const res = await getRecord(row.id)
  currentRecord.value = res.data.data
  resetReviewForm()
  reviewForm.actual_amount = res.data.data.amount
  // Load current attachments for this record
  reviewAttachments.value = []
  reviewUploadingFile.value = false
  try {
    const attRes = await getRecordAttachments(row.id)
    reviewAttachments.value = attRes.data.data || []
  } catch { /* ignore */ }
  reviewDialogVisible.value = true
}

async function handleReviewSubmit() {
  if (!currentRecord.value) return
  const valid = await reviewFormRef.value?.validate().catch(() => false)
  if (!valid) return

  const actionText = reviewForm.action === 'approve' ? '通过' : '驳回'
  await ElMessageBox.confirm(
    `确定要${actionText}记录「${currentRecord.value.record_no}」吗？`,
    '审核确认',
    { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
  )

  reviewLoading.value = true
  try {
    const data: FinanceReviewReq = {
      action: reviewForm.action,
      notes: reviewForm.notes || undefined,
    }
    if (reviewForm.action === 'approve') {
      data.actual_amount = reviewForm.actual_amount
    }
    await reviewRecord(currentRecord.value.id, data)
    ElMessage.success('审核成功')
    reviewDialogVisible.value = false
    fetchList()
  } finally {
    reviewLoading.value = false
  }
}

const reviewAttachments = ref<FinanceAttachmentResp[]>([])
const reviewUploadLoading = ref(false)
const reviewUploadingFile = ref(false)

async function handleReviewUpload(options: { file: File }) {
  if (!currentRecord.value?.id) return
  reviewUploadingFile.value = true
  try {
    const formData = new FormData()
    formData.append('file', options.file)
    await createRecordAttachment(currentRecord.value.id, formData)
    ElMessage.success('附件上传成功')
    // Refresh attachment list
    const res = await getRecordAttachments(currentRecord.value.id)
    reviewAttachments.value = res.data.data || []
  } catch {
    ElMessage.error('上传失败')
  } finally {
    reviewUploadingFile.value = false
  }
}

async function handleDownloadAttachment(row: FinanceAttachmentResp) {
  if (!currentRecord.value?.id) return
  try {
    const res = await downloadRecordAttachment(currentRecord.value.id, row.id)
    const url = res.data.data?.url
    if (url) {
      window.open(url, '_blank')
    }
  } catch {
    ElMessage.error('获取下载链接失败')
  }
}

onMounted(fetchList)
</script>

<template>
  <div class="record-page">
    <el-card shadow="never" class="search-card">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="记录号">
          <el-input v-model="searchForm.record_no" placeholder="请输入记录号" clearable style="width: 200px" />
        </el-form-item>
        <el-form-item label="一级分类">
          <el-select v-model="searchForm.category_l1" placeholder="全部" clearable style="width: 150px" @change="onSearchL1Change">
            <el-option v-for="c in searchL1Options" :key="c.id" :label="c.category_name" :value="c.category_name" />
          </el-select>
        </el-form-item>
        <el-form-item label="二级分类">
          <el-select v-model="searchForm.category_l2" placeholder="全部" clearable :disabled="!searchForm.category_l1" style="width: 150px" @change="onSearchL2Change">
            <el-option v-for="c in searchL2Options" :key="c.id" :label="c.category_name" :value="c.category_name" />
          </el-select>
        </el-form-item>
        <el-form-item label="三级分类">
          <el-select v-model="searchForm.category_l3" placeholder="全部" clearable :disabled="!searchForm.category_l2" style="width: 150px">
            <el-option v-for="c in searchL3Options" :key="c.id" :label="c.category_name" :value="c.category_name" />
          </el-select>
        </el-form-item>
        <el-form-item label="审核状态">
          <el-select v-model="searchForm.review_status" placeholder="全部" clearable style="width: 140px">
            <el-option label="待审核" :value="1" />
            <el-option label="已通过" :value="2" />
            <el-option label="已驳回" :value="3" />
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
        <el-button v-permission="'shop:finance:record:create'" type="primary" @click="openCreateDialog">
          新建记录
        </el-button>
        <el-button v-permission="'shop:finance:record:export'" type="success" @click="handleExport">
          导出
        </el-button>
      </div>

      <el-table v-loading="loading" :data="tableData" border stripe style="width: 100%">
        <el-table-column prop="id" label="序号" width="70" align="center" />
        <el-table-column prop="record_no" label="记录号" min-width="180" />
        <el-table-column prop="account_name" label="账户名" min-width="120" />
        <el-table-column prop="category_l1" label="一级分类" min-width="120" />
        <el-table-column prop="category_l2" label="二级分类" min-width="120" />
        <el-table-column prop="category_l3" label="三级分类" min-width="120" />
        <el-table-column label="类型" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="recordTypeTagType(row.record_type)" size="small">
              {{ recordTypeText(row.record_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="金额" width="120" align="right">
          <template #default="{ row }">
            {{ formatAmount(row.amount) }}
          </template>
        </el-table-column>
        <el-table-column label="实际金额" width="120" align="right">
          <template #default="{ row }">
            {{ formatAmount(row.actual_amount) }}
          </template>
        </el-table-column>
        <el-table-column label="审核状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="reviewStatusTagType(row.review_status)" size="small">
              {{ reviewStatusText(row.review_status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="记账日期" min-width="110">
          <template #default="{ row }">
            {{ row.record_date || '-' }}
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
        <el-table-column label="操作" width="320" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetailDialog(row)">
              详情
            </el-button>
            <el-button
              v-permission="'shop:finance:record:audit'"
              type="success"
              link
              size="small"
              :disabled="row.review_status === 2"
              @click="openReviewDialog(row)"
            >
              审核
            </el-button>
            <el-button
              v-permission="'shop:finance:record:update'"
              type="primary"
              link
              size="small"
              :disabled="row.review_status === 2"
              @click="openEditDialog(row)"
            >
              编辑
            </el-button>
            <el-button
              v-permission="'shop:finance:record:delete'"
              type="danger"
              link
              size="small"
              :disabled="row.review_status === 2"
              @click="handleDelete(row)"
            >
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

    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="640px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="100px" :disabled="isReadonly">
        <el-form-item label="账户" prop="account_id">
          <el-select v-model="formData.account_id" placeholder="请选择账户" filterable style="width: 100%">
            <el-option
              v-for="a in accountOptions"
              :key="a.id"
              :label="a.account_name"
              :value="a.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="分类" prop="category_id">
          <el-space alignment="flex-start" wrap>
            <el-select v-model="formData.category_l1_id" placeholder="一级分类" style="width: 180px" @change="onDialogL1Change">
              <el-option v-for="c in dialogL1Options" :key="c.id" :label="c.category_name" :value="c.id" />
            </el-select>
            <el-select v-model="formData.category_l2_id" placeholder="二级分类" :disabled="!formData.category_l1_id" style="width: 180px" @change="onDialogL2Change">
              <el-option v-for="c in dialogL2Options" :key="c.id" :label="c.category_name" :value="c.id" />
            </el-select>
            <el-select v-model="formData.category_id" placeholder="三级分类" :disabled="!formData.category_l2_id" style="width: 180px" @change="onDialogL3Change">
              <el-option v-for="c in dialogL3Options" :key="c.id" :label="c.category_name" :value="c.id" />
            </el-select>
          </el-space>
        </el-form-item>
        <el-form-item label="金额" prop="amount">
          <el-input-number
            v-model="formData.amount"
            :precision="2"
            :min="0"
            :step="100"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="关联订单">
          <el-select v-model="formData.order_group_id" placeholder="请选择关联订单" filterable clearable style="width: 100%">
            <el-option
              v-for="o in orderOptions"
              :key="o.id"
              :label="o.order_no"
              :value="o.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="记账日期" prop="record_date">
          <el-date-picker
            v-model="formData.record_date"
            type="date"
            value-format="YYYY-MM-DD"
            placeholder="请选择记账日期"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="备注">
          <el-input
            v-model="formData.remark"
            type="textarea"
            :rows="3"
            placeholder="请输入备注"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">{{ isReadonly ? '关闭' : '取消' }}</el-button>
        <el-button v-if="!isReadonly" type="primary" :loading="submitLoading" @click="handleSubmit">
          确定
        </el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="reviewDialogVisible"
      :title="`审核记录 - ${currentRecord?.record_no || ''}`"
      width="480px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-descriptions v-if="currentRecord" :column="1" border size="small" class="review-summary">
        <el-descriptions-item label="记录号">{{ currentRecord.record_no }}</el-descriptions-item>
        <el-descriptions-item label="账户">{{ currentRecord.account_name }}</el-descriptions-item>
          <el-descriptions-item label="一级分类">{{ currentRecord.category_l1 || '-' }}</el-descriptions-item>
          <el-descriptions-item label="二级分类">{{ currentRecord.category_l2 || '-' }}</el-descriptions-item>
          <el-descriptions-item label="三级分类">{{ currentRecord.category_l3 || '-' }}</el-descriptions-item>
        <el-descriptions-item label="金额">{{ formatAmount(currentRecord.amount) }}</el-descriptions-item>
        <el-descriptions-item label="创建人">{{ currentRecord.created_by_name || currentRecord.created_by || '-' }}</el-descriptions-item>
      </el-descriptions>
      <el-form ref="reviewFormRef" :model="reviewForm" :rules="reviewRules" label-width="90px" style="margin-top: 16px">
        <el-form-item label="审核动作" prop="action">
          <el-radio-group v-model="reviewForm.action">
            <el-radio value="approve">通过</el-radio>
            <el-radio value="reject">驳回</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="reviewForm.action === 'approve'" label="实际金额" prop="actual_amount">
          <el-input-number
            v-model="reviewForm.actual_amount"
            :precision="2"
            :min="0"
            :step="100"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="审核备注">
          <el-input
            v-model="reviewForm.notes"
            type="textarea"
            :rows="3"
            placeholder="请输入审核备注"
          />
        </el-form-item>
      </el-form>
      <!-- Attachment upload area -->
      <div style="margin-top: 16px">
        <div style="display: flex; align-items: center; margin-bottom: 8px">
          <span style="font-size: 14px; font-weight: 500; color: #303133">审核凭证</span>
          <span style="font-size: 12px; color: #909399; margin-left: 8px">（可选）</span>
        </div>
        <el-upload
          :http-request="handleReviewUpload"
          :show-file-list="false"
          :disabled="reviewUploadingFile"
          accept=".jpg,.jpeg,.png,.pdf,.doc,.docx,.xls,.xlsx,.zip"
        >
          <el-button size="small" :loading="reviewUploadingFile" type="primary" plain>
            {{ reviewUploadingFile ? '上传中...' : '选择文件' }}
          </el-button>
          <template #tip>
            <div style="font-size: 12px; color: #909399; margin-top: 4px">
              支持 jpg/png/pdf/doc/xls/zip，单个文件不超过 10MB
            </div>
          </template>
        </el-upload>
        <!-- Already uploaded attachments -->
        <div v-if="reviewAttachments.length > 0" style="margin-top: 8px">
          <div v-for="att in reviewAttachments" :key="att.id" style="display: flex; align-items: center; gap: 8px; padding: 4px 0">
            <el-icon><Document /></el-icon>
            <span style="font-size: 13px; flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ att.file_name }}</span>
            <span style="font-size: 12px; color: #909399">{{ formatFileSize(att.file_size) }}</span>
            <el-button type="primary" link size="small" @click="handleDownloadAttachment(att)">下载</el-button>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="reviewDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="reviewLoading" @click="handleReviewSubmit">提交</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.record-page { padding: 16px; }
.search-card { margin-bottom: 16px; }
.search-card :deep(.el-card__body) { padding-bottom: 2px; }
.table-header { margin-bottom: 16px; display: flex; gap: 8px; }
.pagination-wrap { display: flex; justify-content: flex-end; margin-top: 16px; }
.review-summary { margin-top: 8px; }
</style>