<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import type { UploadFile, UploadRawFile } from 'element-plus'
import { Document, Plus } from '@element-plus/icons-vue'
import {
  getOrderList,
  getOrder,
  createOrder,
  cancelOrder,
  cancelOrderItem,
  exportOrders,
  getItemWorkflowLogs,
  advanceItemWorkflow,
  getItemAttachments,
  createItemAttachment,
  getItemAttachmentDownloadURL,
} from '@/api/shop/order'
import { getCustomerList } from '@/api/shop/customer'
import { getShopProducts } from '@/api/shop/product'
import type {
  OrderResp,
  OrderItemResp,
  OrderCreateReq,
  OrderWorkflowLogResp,
  OrderAttachmentResp,
  ShopCustomerResp,
  ShopProductResp,
} from '@/types/system'

const loading = ref(false)
const tableData = ref<OrderResp[]>([])
const total = ref(0)

const searchForm = reactive({
  order_no: '',
  order_status: null as number | null,
})

const pagination = reactive({ page: 1, page_size: 10 })

async function fetchList() {
  loading.value = true
  try {
    const res = await getOrderList({
      page: pagination.page,
      page_size: pagination.page_size,
      order_no: searchForm.order_no || undefined,
      order_status: searchForm.order_status ?? undefined,
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
  searchForm.order_no = ''
  searchForm.order_status = null
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

const orderStatusTagType = (s: number) => {
  if (s === 1) return 'info'
  if (s === 2) return 'warning'
  if (s === 3) return 'success'
  return 'danger'
}
const orderStatusText = (s: number) => {
  if (s === 1) return '待处理'
  if (s === 2) return '进行中'
  if (s === 3) return '已完成'
  return '已取消'
}
const itemStatusTagType = (s: number) => orderStatusTagType(s)
const itemStatusText = (s: number) => orderStatusText(s)

function formatTime(val: string | null): string {
  if (!val) return '-'
  return val.replace('T', ' ').substring(0, 19)
}
function formatAmount(val: number): string {
  return Number(val).toFixed(2)
}

const detailDialogVisible = ref(false)
const detailLoading = ref(false)
const currentOrder = ref<OrderResp | null>(null)
type DetailTab = 'items' | 'logs'
const detailTab = ref<DetailTab>('items')
const detailLogsLoading = ref(false)
const detailLogs = ref<(OrderWorkflowLogResp & { product_name: string })[]>([])

async function openDetailDialog(row: OrderResp) {
  detailDialogVisible.value = true
  detailLoading.value = true
  detailTab.value = 'items'
  detailLogs.value = []
  try {
    const res = await getOrder(row.id)
    currentOrder.value = res.data.data
  } finally {
    detailLoading.value = false
  }
}

async function refreshDetail() {
  if (!currentOrder.value) return
  const res = await getOrder(currentOrder.value.id)
  currentOrder.value = res.data.data
  detailLogs.value = []
}

async function loadDetailLogs() {
  if (!currentOrder.value || !currentOrder.value.items) return
  if (detailLogs.value.length > 0) return
  detailLogsLoading.value = true
  try {
    const all: (OrderWorkflowLogResp & { product_name: string })[] = []
    attachmentsMap.value = {}
    for (const item of currentOrder.value.items) {
      const [logRes, attRes] = await Promise.all([
        getItemWorkflowLogs(currentOrder.value.id, item.id),
        getItemAttachments(currentOrder.value.id, item.id),
      ])
      const logs = logRes.data.data || []
      logs.forEach((log) => {
        all.push({ ...log, order_item_id: item.id, product_name: item.product_name })
      })
      attachmentsMap.value[item.id] = attRes.data.data || []
    }
    all.sort((a, b) => (a.operated_at < b.operated_at ? 1 : -1))
    detailLogs.value = all
  } finally {
    detailLogsLoading.value = false
  }
}

const attachmentsMap = ref<Record<number, OrderAttachmentResp[]>>({})

function getLogAttachments(itemId: number, workflowLogId: number | undefined): OrderAttachmentResp[] {
  if (!workflowLogId) return []
  const list = attachmentsMap.value[itemId] || []
  return list.filter((a) => a.workflow_log_id === workflowLogId)
}

async function handleDownloadAttachment(itemId: number, attId: number) {
  if (!currentOrder.value) return
  const orderId = currentOrder.value.id
  const res = await getItemAttachmentDownloadURL(orderId, itemId, attId)
  const url = res.data.data?.url
  if (!url) {
    ElMessage.error('获取下载链接失败')
    return
  }
  window.open(url, '_blank')
}

function handleDetailTabClick(tab: { props?: { name?: string } }) {
  if (tab?.props?.name === 'logs') {
    loadDetailLogs()
  }
}

const advanceDialogVisible = ref(false)
const advanceLoading = ref(false)
const advanceForm = reactive({ notes: '' })
const advanceAttachments = ref<UploadFile[]>([])
const advanceContext = ref<{ orderId: number; item: OrderItemResp } | null>(null)
const ADVANCE_FILE_INPUT_ID = 'advance-file-input'

function openAdvanceDialog(item: OrderItemResp) {
  if (!currentOrder.value) return
  advanceContext.value = { orderId: currentOrder.value.id, item }
  advanceForm.notes = ''
  advanceAttachments.value = []
  // Reset native input so picking the same file twice fires @change
  const input = document.getElementById(ADVANCE_FILE_INPUT_ID) as HTMLInputElement | null
  if (input) input.value = ''
  advanceDialogVisible.value = true
}

function handleAdvanceFileInput(e: Event) {
  const input = e.target as HTMLInputElement
  const files = input.files
  if (!files || files.length === 0) return
  const max = 20 * 1024 * 1024
  for (const file of Array.from(files)) {
    if (file.size > max) {
      ElMessage.error(`文件 ${file.name} 超过 20MB，已跳过`)
      continue
    }
    advanceAttachments.value.push({
      uid: Date.now() + Math.random(),
      name: file.name,
      size: file.size,
      status: 'ready',
      raw: file,
    } as UploadFile)
  }
  // Reset so picking same file again fires @change
  input.value = ''
}

function removeAdvanceAttachment(uid: number) {
  advanceAttachments.value = advanceAttachments.value.filter(f => f.uid !== uid)
}

async function handleAdvanceSubmit() {
  if (!advanceContext.value) return
  if (!advanceForm.notes.trim()) {
    ElMessage.warning('请填写备注')
    return
  }
  advanceLoading.value = true
  try {
    const { orderId, item } = advanceContext.value
    const advanceRes = await advanceItemWorkflow(orderId, item.id, { notes: advanceForm.notes })
    const workflowLogId = advanceRes.data.data?.workflow_log_id

    let uploaded = 0
    for (const f of advanceAttachments.value) {
      const raw = f.raw as UploadRawFile | undefined
      if (!raw) continue
      const fd = new FormData()
      fd.append('file', raw)
      if (workflowLogId) fd.append('workflow_log_id', String(workflowLogId))
      try {
        await createItemAttachment(orderId, item.id, fd)
        uploaded++
      } catch (e) {
        ElMessage.error(`附件 ${f.name} 上传失败`)
      }
    }
    if (uploaded > 0) ElMessage.success(`已上传 ${uploaded} 个附件`)
    ElMessage.success('推进成功')
    advanceDialogVisible.value = false
    await refreshDetail()
  } finally {
    advanceLoading.value = false
  }
}

function formatFileSize(size: number): string {
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  return `${(size / 1024 / 1024).toFixed(2)} MB`
}

async function handleCancelItem(item: OrderItemResp) {
  if (!currentOrder.value) return
  await ElMessageBox.confirm(
    `确定要取消明细 ${item.product_name} 吗？`,
    '取消确认',
    { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
  )
  await cancelOrderItem(currentOrder.value.id, item.id)
  ElMessage.success('取消明细成功')
  await refreshDetail()
}

async function handleCancelOrder(row: OrderResp) {
  await ElMessageBox.confirm(
    `确定要取消订单 ${row.order_no} 吗？`,
    '取消确认',
    { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
  )
  await cancelOrder(row.id)
  ElMessage.success('取消订单成功')
  fetchList()
}

async function handleExport() {
  const res = await exportOrders({
    order_no: searchForm.order_no || undefined,
    order_status: searchForm.order_status ?? undefined,
  })
  const data = res.data.data
  if (!data || data.length === 0) {
    ElMessage.warning('没有数据可导出')
    return
  }
  const headers = ['订单号', '客户', '金额', '商品数', '状态', '创建人', '创建时间']
  const rows = data.map((o) => [
    o.order_no,
    o.customer_name,
    formatAmount(o.total_amount),
    o.item_count,
    orderStatusText(o.order_status),
    o.created_by_name || String(o.created_by),
    formatTime(o.created_at),
  ])
  const csv = [headers, ...rows]
    .map((r) => r.map((cell) => `"${String(cell).replace(/"/g, '""')}"`).join(','))
    .join('\n')
  const blob = new Blob(['\uFEFF' + csv], { type: 'text/csv;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `订单列表_${new Date().toISOString().slice(0, 10)}.csv`
  link.click()
  URL.revokeObjectURL(url)
  ElMessage.success('导出成功')
}

const createDialogVisible = ref(false)
const createSubmitLoading = ref(false)
const createFormRef = ref<FormInstance>()
const customerOptions = ref<ShopCustomerResp[]>([])
const createForm = reactive({
  customer_id: undefined as number | undefined,
  remark: '',
})

const createFormRules = reactive<FormRules>({
  customer_id: [{ required: true, message: '请选择客户', trigger: 'change' }],
})

interface SelectedItem {
  shop_product_id: number
  product_name: string
  unit_price: number
  quantity: number
}

const selectedItems = ref<SelectedItem[]>([])

const totalAmount = computed(() => {
  return selectedItems.value
    .reduce((sum, it) => sum + Number(it.unit_price) * Number(it.quantity), 0)
    .toFixed(2)
})

async function loadCustomers() {
  const res = await getCustomerList({ page: 1, page_size: 1000 })
  customerOptions.value = res.data.data.list
}

function openCreateDialog() {
  createDialogVisible.value = true
  Object.assign(createForm, { customer_id: undefined, remark: '' })
  selectedItems.value = []
  loadCustomers()
}

function addItemFromProduct(p: ShopProductResp) {
  const existing = selectedItems.value.find((it) => it.shop_product_id === p.id)
  if (existing) {
    existing.quantity += 1
  } else {
    selectedItems.value.push({
      shop_product_id: p.id,
      product_name: p.product_name,
      unit_price: p.shop_price,
      quantity: 1,
    })
  }
}

function removeSelectedItem(index: number) {
  selectedItems.value.splice(index, 1)
}

const pickerVisible = ref(false)
const pickerLoading = ref(false)
const pickerProducts = ref<ShopProductResp[]>([])
const pickerKeyword = ref('')
const pickerTableRef = ref<any>()

async function openPicker() {
  pickerVisible.value = true
  pickerKeyword.value = ''
  pickerLoading.value = true
  try {
    const res = await getShopProducts({ page: 1, page_size: 1000, status: 1 })
    pickerProducts.value = res.data.data.list
  } finally {
    pickerLoading.value = false
  }
}

const filteredPickerProducts = computed(() => {
  if (!pickerKeyword.value) return pickerProducts.value
  const kw = pickerKeyword.value.toLowerCase()
  return pickerProducts.value.filter(
    (p) =>
      p.product_name.toLowerCase().includes(kw) ||
      p.product_code.toLowerCase().includes(kw)
  )
})

function handlePickerConfirm() {
  const rows = pickerTableRef.value?.getSelectionRows() as ShopProductResp[] | undefined
  if (!rows || rows.length === 0) {
    ElMessage.warning('请选择商品')
    return
  }
  rows.forEach(addItemFromProduct)
  pickerVisible.value = false
  ElMessage.success(`已添加 ${rows.length} 个商品`)
}

async function handleCreateSubmit() {
  const valid = await createFormRef.value?.validate().catch(() => false)
  if (!valid) return
  if (selectedItems.value.length === 0) {
    ElMessage.warning('请至少选择一个商品')
    return
  }
  createSubmitLoading.value = true
  try {
    const data: OrderCreateReq = {
      customer_id: createForm.customer_id as number,
      remark: createForm.remark || undefined,
      items: selectedItems.value.map((it) => ({
        shop_product_id: it.shop_product_id,
        quantity: it.quantity,
      })),
    }
    await createOrder(data)
    ElMessage.success('创建成功')
    createDialogVisible.value = false
    fetchList()
  } finally {
    createSubmitLoading.value = false
  }
}

onMounted(() => {
  fetchList()
})
</script>

<template>
  <div class="order-page">
    <el-card shadow="never" class="search-card">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="订单号">
          <el-input v-model="searchForm.order_no" placeholder="请输入订单号" clearable />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.order_status" placeholder="全部" clearable style="width: 140px">
            <el-option label="待处理" :value="1" />
            <el-option label="进行中" :value="2" />
            <el-option label="已完成" :value="3" />
            <el-option label="已取消" :value="4" />
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
        <el-button v-permission="'shop:order:create'" type="primary" @click="openCreateDialog">
          新建订单
        </el-button>
        <el-button v-permission="'shop:order:export'" type="success" @click="handleExport">
          导出
        </el-button>
      </div>

      <el-table v-loading="loading" :data="tableData" border stripe style="width: 100%">
        <el-table-column prop="id" label="序号" width="70" align="center" />
        <el-table-column prop="order_no" label="订单号" min-width="180" />
        <el-table-column prop="customer_name" label="客户名称" min-width="120" />
        <el-table-column label="金额" width="120" align="right">
          <template #default="{ row }">
            {{ formatAmount(row.total_amount) }}
          </template>
        </el-table-column>
        <el-table-column prop="item_count" label="商品数" width="80" align="center" />
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="orderStatusTagType(row.order_status)" size="small">
              {{ orderStatusText(row.order_status) }}
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
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetailDialog(row)">
              详情
            </el-button>
            <el-button
              v-permission="'shop:order:cancel'"
              type="danger"
              link
              size="small"
              :disabled="row.order_status !== 1"
              @click="handleCancelOrder(row)"
            >
              取消
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
      v-model="detailDialogVisible"
      :title="`订单详情 - ${currentOrder?.order_no || ''}`"
      width="800px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <div v-loading="detailLoading">
        <el-card v-if="currentOrder" shadow="never" class="detail-header">
          <el-descriptions :column="2" border size="small">
            <el-descriptions-item label="订单号">{{ currentOrder.order_no }}</el-descriptions-item>
            <el-descriptions-item label="客户">{{ currentOrder.customer_name }}</el-descriptions-item>
            <el-descriptions-item label="金额">{{ formatAmount(currentOrder.total_amount) }}</el-descriptions-item>
            <el-descriptions-item label="状态">
              <el-tag :type="orderStatusTagType(currentOrder.order_status)" size="small">
                {{ orderStatusText(currentOrder.order_status) }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="备注" :span="2">
              {{ currentOrder.remark || '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="创建人">
              {{ currentOrder.created_by_name || currentOrder.created_by || '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="创建时间">{{ formatTime(currentOrder.created_at) }}</el-descriptions-item>
          </el-descriptions>
        </el-card>

        <el-tabs v-if="currentOrder" v-model="detailTab" class="detail-tabs" @tab-click="handleDetailTabClick">
          <el-tab-pane label="商品明细" name="items">
            <el-table
              v-if="currentOrder.items && currentOrder.items.length > 0"
              :data="currentOrder.items"
              border
              stripe
              style="width: 100%"
            >
              <el-table-column prop="product_name" label="商品名" min-width="160" />
              <el-table-column prop="quantity" label="数量" width="80" align="center" />
              <el-table-column label="单价" width="100" align="right">
                <template #default="{ row }">
                  {{ formatAmount(row.unit_price) }}
                </template>
              </el-table-column>
              <el-table-column label="小计" width="120" align="right">
                <template #default="{ row }">
                  {{ formatAmount(row.total_price) }}
                </template>
              </el-table-column>
              <el-table-column label="当前节点" min-width="120">
                <template #default="{ row }">
                  {{ row.current_node_name || '-' }}
                </template>
              </el-table-column>
              <el-table-column label="状态" width="100" align="center">
                <template #default="{ row }">
                  <el-tag :type="itemStatusTagType(row.item_status)" size="small">
                    {{ itemStatusText(row.item_status) }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="操作" width="200" fixed="right">
                <template #default="{ row }">
                  <el-button
                    v-permission="'shop:order:advance'"
                    type="primary"
                    link
                    size="small"
                    :disabled="!(row.item_status === 1 || row.item_status === 2) || !row.next_node_name"
                    @click="openAdvanceDialog(row)"
                  >
                    推进流程
                  </el-button>
                  <el-button
                    v-permission="'shop:order:cancel'"
                    type="danger"
                    link
                    size="small"
                    :disabled="!(row.item_status === 1 || row.item_status === 2)"
                    @click="handleCancelItem(row)"
                  >
                    取消明细
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
            <el-empty v-else description="暂无明细" />
          </el-tab-pane>
          <el-tab-pane label="流程跟进" name="logs">
            <div v-loading="detailLogsLoading">
              <el-table
                v-if="detailLogs.length > 0"
                :data="detailLogs"
                border
                stripe
                style="width: 100%"
              >
                <el-table-column label="时间" min-width="170">
                  <template #default="{ row }">
                    {{ formatTime(row.operated_at) }}
                  </template>
                </el-table-column>
                <el-table-column prop="product_name" label="商品" min-width="140" />
                <el-table-column label="节点" min-width="160">
                  <template #default="{ row }">
                    <el-tag size="small">第{{ row.node_index }}步 · {{ row.node_name }}</el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="notes" label="备注" min-width="180" show-overflow-tooltip />
                <el-table-column label="操作人" min-width="120">
                  <template #default="{ row }">
                    {{ row.operator_name || row.operator_id || '-' }}
                  </template>
                </el-table-column>
                <el-table-column label="附件" min-width="220">
                  <template #default="{ row }">
                    <template v-if="row.order_item_id && getLogAttachments(row.order_item_id, row.id).length > 0">
                      <el-link
                        v-for="att in getLogAttachments(row.order_item_id, row.id)"
                        :key="att.id"
                        type="primary"
                        :underline="false"
                        style="margin-right: 8px"
                        @click="handleDownloadAttachment(row.order_item_id, att.id)"
                      >
                        <el-icon style="vertical-align: middle"><Document /></el-icon>
                        <span style="margin-left: 2px">{{ att.file_name }}</span>
                      </el-link>
                    </template>
                    <span v-else>-</span>
                  </template>
                </el-table-column>
              </el-table>
              <el-empty v-else description="暂无跟进记录" />
            </div>
          </el-tab-pane>
        </el-tabs>
      </div>
      <template #footer>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="advanceDialogVisible"
      title="推进流程"
      width="520px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-form v-if="advanceContext" label-width="90px">
        <el-form-item label="当前节点">
          <el-tag size="small">{{ advanceContext.item.current_node_name || '-' }}</el-tag>
        </el-form-item>
        <el-form-item label="下一节点">
          <el-tag size="small" type="success">{{ advanceContext.item.next_node_name || '已是最后节点' }}</el-tag>
        </el-form-item>
        <el-form-item label="备注">
          <el-input
            v-model="advanceForm.notes"
            type="textarea"
            :rows="4"
            placeholder="请填写本节点服务备注"
          />
        </el-form-item>
        <el-form-item label="附件">
          <!-- Native <label for> triggers file picker at browser level — no JS
               click forwarding, no ref binding, no dialog stacking issue.         -->
          <input
            :id="ADVANCE_FILE_INPUT_ID"
            type="file"
            multiple
            accept="*/*"
            class="hidden-file-input"
            @change="handleAdvanceFileInput"
          />
          <label :for="ADVANCE_FILE_INPUT_ID" class="el-button el-button--primary is-plain">
            <el-icon style="margin-right: 4px;"><Plus /></el-icon>
            选择文件
          </label>
          <span class="el-upload__tip" style="margin-left: 12px;">可上传多个文件，单个不超过 20MB（随流程记录一起提交）</span>
          <div v-if="advanceAttachments.length" class="upload-list-preview">
            <div v-for="f in advanceAttachments" :key="f.uid" class="upload-item">
              <span class="upload-item-name">{{ f.name }}</span>
              <span class="upload-item-size">{{ formatFileSize(f.size || 0) }}</span>
              <el-button type="danger" link size="small" @click="removeAdvanceAttachment(f.uid)">移除</el-button>
            </div>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="advanceDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="advanceLoading" @click="handleAdvanceSubmit">推进</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="createDialogVisible"
      title="新建订单"
      width="720px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-form ref="createFormRef" :model="createForm" :rules="createFormRules" label-width="90px">
        <el-form-item label="客户" prop="customer_id">
          <el-select v-model="createForm.customer_id" placeholder="请选择客户" filterable style="width: 100%">
            <el-option
              v-for="c in customerOptions"
              :key="c.id"
              :label="c.customer_name"
              :value="c.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="备注">
          <el-input
            v-model="createForm.remark"
            type="textarea"
            :rows="2"
            placeholder="请输入备注"
          />
        </el-form-item>
        <el-form-item label="选商品">
          <el-button type="primary" plain @click="openPicker">添加商品</el-button>
        </el-form-item>
        <el-form-item label="商品明细">
          <el-table
            v-if="selectedItems.length > 0"
            :data="selectedItems"
            border
            style="width: 100%"
          >
            <el-table-column prop="product_name" label="商品名" min-width="180" />
            <el-table-column label="数量" width="160" align="center">
              <template #default="{ row }">
                <el-input-number
                  v-model="row.quantity"
                  :min="1"
                  :max="999"
                  size="small"
                  style="width: 120px"
                />
              </template>
            </el-table-column>
            <el-table-column label="单价" width="100" align="right">
              <template #default="{ row }">
                {{ formatAmount(row.unit_price) }}
              </template>
            </el-table-column>
            <el-table-column label="小计" width="100" align="right">
              <template #default="{ row }">
                {{ formatAmount(Number(row.unit_price) * Number(row.quantity)) }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="80" align="center" fixed="right">
              <template #default="{ $index }">
                <el-button type="danger" link size="small" @click="removeSelectedItem($index)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-else description="尚未选择商品" />
        </el-form-item>
        <el-form-item label="订单合计">
          <span class="total-amount">￥{{ totalAmount }}</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="createSubmitLoading" @click="handleCreateSubmit">提交</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="pickerVisible"
      title="选择商品"
      width="720px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <div v-loading="pickerLoading">
        <div class="picker-search">
          <el-input
            v-model="pickerKeyword"
            placeholder="搜索商品名称/编号"
            clearable
            style="width: 260px"
          />
        </div>
        <el-table
          ref="pickerTableRef"
          :data="filteredPickerProducts"
          border
          max-height="400"
          style="width: 100%"
        >
          <el-table-column type="selection" width="50" />
          <el-table-column prop="id" label="编号" width="80" align="center" />
          <el-table-column prop="product_name" label="名称" min-width="200" />
          <el-table-column label="售价" width="120" align="right">
            <template #default="{ row }">
              {{ formatAmount(row.shop_price) }}
            </template>
          </el-table-column>
        </el-table>
      </div>
      <template #footer>
        <el-button @click="pickerVisible = false">取消</el-button>
        <el-button type="primary" @click="handlePickerConfirm">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.order-page { padding: 16px; }
.search-card { margin-bottom: 16px; }
.search-card :deep(.el-card__body) { padding-bottom: 2px; }
.table-header { margin-bottom: 16px; display: flex; gap: 8px; }
.pagination-wrap { display: flex; justify-content: flex-end; margin-top: 16px; }
.detail-header { margin-bottom: 16px; }
.detail-tabs { margin-top: 8px; }
.total-amount { font-size: 18px; font-weight: 600; color: #f56c6c; }
.picker-search { margin-bottom: 12px; }
.upload-list-preview { margin-top: 8px; }
.upload-item { display: flex; justify-content: space-between; padding: 4px 0; font-size: 13px; color: #606266; }
.upload-item-name { color: #409eff; }
.upload-item-size { color: #909399; }
/* Hidden file input — triggered by <label for> at browser level, no JS needed */
.hidden-file-input {
  position: absolute;
  width: 0;
  height: 0;
  opacity: 0;
  pointer-events: none;
}
</style>
