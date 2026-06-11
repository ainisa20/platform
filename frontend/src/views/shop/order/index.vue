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
  exportOrders,
  getItemWorkflow,
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
  OrderWorkflowNodeResp,
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
const expandedItemId = ref<number | null>(null)
const expandLoading = ref(false)

interface ExpandData {
  nodes: OrderWorkflowNodeResp[]
  currentIndex: number
  logs: OrderWorkflowLogResp[]
  attachments: OrderAttachmentResp[]
}
const expandDataMap = ref<Record<number, ExpandData>>({})

async function openDetailDialog(row: OrderResp) {
  detailDialogVisible.value = true
  detailLoading.value = true
  expandedItemId.value = null
  expandDataMap.value = {}
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
}

function handlePrintOrder() {
  window.print()
}

function getLogAttachments(itemId: number, workflowLogId: number | undefined): OrderAttachmentResp[] {
  if (!workflowLogId) return []
  const data = expandDataMap.value[itemId]
  if (!data) return []
  return data.attachments.filter((a: OrderAttachmentResp) => a.workflow_log_id === workflowLogId)
}

function getNodeLog(itemId: number, nodeIndex: number): OrderWorkflowLogResp | undefined {
  const data = expandDataMap.value[itemId]
  if (!data) return undefined
  return data.logs.find((l: OrderWorkflowLogResp) => l.node_index === nodeIndex)
}

async function openInlineAdvance(item: OrderItemResp) {
  if (!currentOrder.value) return
  if (expandedItemId.value === item.id) {
    expandedItemId.value = null
    return
  }
  expandedItemId.value = item.id
  advanceForm.notes = ''
  advanceAttachments.value = []
  advanceContext.value = { orderId: currentOrder.value.id, item }
  const input = document.getElementById(ADVANCE_FILE_INPUT_ID) as HTMLInputElement | null
  if (input) input.value = ''
  expandLoading.value = true
  try {
    const orderId = currentOrder.value.id
    const [wfRes, logRes, attRes] = await Promise.all([
      getItemWorkflow(orderId, item.id),
      getItemWorkflowLogs(orderId, item.id),
      getItemAttachments(orderId, item.id),
    ])
    const wfData = wfRes.data.data as any
    expandDataMap.value[item.id] = {
      nodes: wfData?.nodes || [],
      currentIndex: wfData?.current_node_index ?? -1,
      logs: logRes.data.data || [],
      attachments: attRes.data.data || [],
    }
  } finally {
    expandLoading.value = false
  }
}

function getCurrentNodeDetail(itemId: number): OrderWorkflowLogResp | undefined {
  const data = expandDataMap.value[itemId]
  if (!data) return undefined
  const idx = data.currentIndex === -1 ? 0 : data.currentIndex
  return data.logs.find((l: OrderWorkflowLogResp) => l.node_index === idx)
}

function getCompletedNodeLog(itemId: number, nodeIndex: number, currentIndex: number): OrderWorkflowLogResp | undefined {
  if (nodeIndex > currentIndex) return undefined
  const data = expandDataMap.value[itemId]
  if (!data) return undefined
  return data.logs.find((l: OrderWorkflowLogResp) => l.node_index === nodeIndex)
}

function getNodeAttachments(itemId: number, workflowLogId: number | undefined): OrderAttachmentResp[] {
  if (!workflowLogId) return []
  const data = expandDataMap.value[itemId]
  if (!data) return []
  return data.attachments.filter((a: OrderAttachmentResp) => a.workflow_log_id === workflowLogId)
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

const advanceLoading = ref(false)
const advanceForm = reactive({ notes: '' })
const advanceAttachments = ref<UploadFile[]>([])
const advanceContext = ref<{ orderId: number; item: OrderItemResp } | null>(null)
const ADVANCE_FILE_INPUT_ID = 'advance-file-input'

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
    advanceForm.notes = ''
    advanceAttachments.value = []
    await refreshDetail()
    const updatedItem = currentOrder.value?.items?.find(i => i.id === item.id)
    if (updatedItem && (updatedItem.item_status === 1 || updatedItem.item_status === 2)) {
      advanceContext.value = { orderId, item: updatedItem }
      const [wfRes, logRes, attRes] = await Promise.all([
        getItemWorkflow(orderId, updatedItem.id),
        getItemWorkflowLogs(orderId, updatedItem.id),
        getItemAttachments(orderId, updatedItem.id),
      ])
      const wfData = wfRes.data.data as any
      expandDataMap.value[updatedItem.id] = {
        nodes: wfData?.nodes || [],
        currentIndex: wfData?.current_node_index ?? -1,
        logs: logRes.data.data || [],
        attachments: attRes.data.data || [],
      }
    } else {
      expandedItemId.value = null
    }
  } finally {
    advanceLoading.value = false
  }
}

function formatFileSize(size: number): string {
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  return `${(size / 1024 / 1024).toFixed(2)} MB`
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
      width="900px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <div v-loading="detailLoading" id="order-print-area">
        <el-card v-if="currentOrder" shadow="never" class="detail-header">
          <el-descriptions :column="2" border size="small">
            <el-descriptions-item label="订单号">{{ currentOrder.order_no }}</el-descriptions-item>
            <el-descriptions-item label="客户">{{ currentOrder.customer_name }}</el-descriptions-item>
            <el-descriptions-item label="金额">
              <span style="font-weight: 600; color: #f56c6c">{{ formatAmount(currentOrder.total_amount) }}</span>
            </el-descriptions-item>
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

        <div v-if="currentOrder && currentOrder.items && currentOrder.items.length > 0" style="margin-top: 20px">
          <div class="product-section-title">商品服务明细</div>
          <div v-for="item in currentOrder.items" :key="item.id" class="product-card">
            <div class="product-card-header">
              <div class="product-info">
                <span class="product-name">{{ item.product_name }}</span>
                <span class="product-meta">x{{ item.quantity }} · 单价 ¥{{ formatAmount(item.unit_price) }} · 小计 ¥{{ formatAmount(item.total_price) }}</span>
              </div>
              <div class="product-actions">
                <el-tag :type="itemStatusTagType(item.item_status)" size="small">
                  {{ itemStatusText(item.item_status) }}
                </el-tag>
                <el-button
                  v-if="item.item_status === 1 || item.item_status === 2"
                  v-permission="'shop:order:advance'"
                  type="primary"
                  size="small"
                  plain
                  @click="openInlineAdvance(item)"
                >
                  {{ expandedItemId === item.id ? '收起' : '推进流程' }}
                </el-button>
              </div>
            </div>

            <div v-if="expandedItemId === item.id" class="workflow-section" v-loading="expandLoading">
              <template v-if="expandDataMap[item.id]">
                <el-steps
                  :active="expandDataMap[item.id].currentIndex === -1 ? 0 : expandDataMap[item.id].currentIndex"
                  finish-status="success"
                  align-center
                  class="workflow-steps"
                >
                  <el-step
                    v-for="node in expandDataMap[item.id].nodes"
                    :key="node.node_index"
                    :title="node.node_name"
                  />
                </el-steps>

                <div class="completed-nodes-detail">
                  <template v-for="node in expandDataMap[item.id].nodes" :key="'d_'+node.node_index">
                    <div
                      v-if="getCompletedNodeLog(item.id, node.node_index, expandDataMap[item.id].currentIndex)"
                      class="completed-node-item"
                    >
                      <div class="completed-node-header">
                        <span class="completed-node-title">第{{ node.node_index }}步 · {{ node.node_name }}</span>
                        <el-tag size="small" type="success">已完成</el-tag>
                      </div>
                      <div class="detail-row">
                        <span class="detail-label">操作人:</span>
                        <span>{{ getCompletedNodeLog(item.id, node.node_index, expandDataMap[item.id].currentIndex)?.operator_name || '-' }}</span>
                      </div>
                      <div class="detail-row">
                        <span class="detail-label">操作时间:</span>
                        <span>{{ formatTime(getCompletedNodeLog(item.id, node.node_index, expandDataMap[item.id].currentIndex)?.operated_at ?? null) }}</span>
                      </div>
                      <div v-if="getCompletedNodeLog(item.id, node.node_index, expandDataMap[item.id].currentIndex)?.notes" class="detail-notes">
                        {{ getCompletedNodeLog(item.id, node.node_index, expandDataMap[item.id].currentIndex)?.notes }}
                      </div>
                      <div v-if="getNodeAttachments(item.id, getCompletedNodeLog(item.id, node.node_index, expandDataMap[item.id].currentIndex)?.id).length" class="detail-attachments">
                        <el-link
                          v-for="att in getNodeAttachments(item.id, getCompletedNodeLog(item.id, node.node_index, expandDataMap[item.id].currentIndex)?.id)"
                          :key="att.id"
                          type="primary"
                          :underline="false"
                          style="margin-right: 12px"
                          @click="handleDownloadAttachment(item.id, att.id)"
                        >
                          <el-icon style="vertical-align: middle"><Document /></el-icon>
                          <span style="margin-left: 2px">{{ att.file_name }}</span>
                        </el-link>
                      </div>
                    </div>
                  </template>
                </div>

                <div
                  v-if="item.item_status === 1 || item.item_status === 2"
                  class="inline-advance-form"
                >
                  <div class="form-title">推进当前节点</div>
                  <div class="form-row">
                    <span class="form-label">当前:</span>
                    <el-tag size="small">{{ item.current_node_name || '起始' }}</el-tag>
                    <span style="margin: 0 8px">→</span>
                    <span class="form-label">下一步:</span>
                    <el-tag size="small" type="success">{{ item.next_node_name || '完成' }}</el-tag>
                  </div>
                  <el-input
                    v-model="advanceForm.notes"
                    type="textarea"
                    :rows="3"
                    placeholder="请填写本节点服务备注"
                    style="margin-top: 12px"
                  />
                  <div class="form-upload-row">
                    <input
                      :id="ADVANCE_FILE_INPUT_ID"
                      type="file"
                      multiple
                      accept="*/*"
                      class="hidden-file-input"
                      @change="handleAdvanceFileInput"
                    />
                    <label :for="ADVANCE_FILE_INPUT_ID" class="el-button el-button--primary is-plain" style="display: inline-flex; align-items: center; cursor: pointer">
                      <el-icon style="margin-right: 4px;"><Plus /></el-icon>
                      选择文件
                    </label>
                    <span style="margin-left: 8px; font-size: 12px; color: #909399">可上传多个，单个不超过 20MB</span>
                  </div>
                  <div v-if="advanceAttachments.length" class="upload-list-preview">
                    <div v-for="f in advanceAttachments" :key="f.uid" class="upload-item">
                      <span class="upload-item-name">{{ f.name }}</span>
                      <span class="upload-item-size">{{ formatFileSize(f.size || 0) }}</span>
                      <el-button type="danger" link size="small" @click="removeAdvanceAttachment(f.uid)">移除</el-button>
                    </div>
                  </div>
                  <div style="text-align: right; margin-top: 12px">
                    <el-button type="primary" :loading="advanceLoading" @click="handleAdvanceSubmit">提交推进</el-button>
                  </div>
                </div>
              </template>
            </div>
          </div>
        </div>
        <el-empty v-else-if="!detailLoading && currentOrder" description="暂无商品明细" />
      </div>
      <template #footer>
        <el-button @click="handlePrintOrder">打印</el-button>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
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
.detail-header { margin-bottom: 0; }
.total-amount { font-size: 18px; font-weight: 600; color: #f56c6c; }
.picker-search { margin-bottom: 12px; }

.product-section-title {
  font-size: 15px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 16px;
  padding-left: 8px;
  border-left: 3px solid var(--el-color-primary);
}

.product-card {
  border: 1px solid #ebeef5;
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 16px;
  background: #fff;
  transition: box-shadow 0.2s;
}
.product-card:hover {
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
}

.product-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.product-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.product-name {
  font-size: 15px;
  font-weight: 600;
  color: #303133;
}
.product-meta {
  font-size: 13px;
  color: #909399;
}
.product-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.workflow-section {
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid #ebeef5;
}

.workflow-steps {
  margin-bottom: 20px;
}
.workflow-steps :deep(.el-step__title) {
  font-size: 13px;
}

.completed-nodes-detail {
  margin-bottom: 16px;
}
.completed-node-item {
  background: #f0f9eb;
  border: 1px solid #e1f3d8;
  border-radius: 6px;
  padding: 10px 14px;
  margin-bottom: 8px;
}
.completed-node-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 6px;
}
.completed-node-title {
  font-size: 13px;
  font-weight: 600;
  color: #67c23a;
}

.current-node-detail {
  background: #f5f7fa;
  border-radius: 6px;
  padding: 12px 16px;
  margin-bottom: 16px;
}
.detail-row {
  font-size: 13px;
  color: #606266;
  margin-bottom: 4px;
}
.detail-label {
  color: #909399;
  margin-right: 4px;
}
.detail-notes {
  font-size: 13px;
  color: #303133;
  margin-top: 8px;
  padding: 6px 10px;
  background: #fff;
  border-radius: 4px;
  border: 1px solid #ebeef5;
}
.detail-attachments {
  margin-top: 8px;
}

.inline-advance-form {
  background: #ecf5ff;
  border-radius: 6px;
  padding: 16px;
  border: 1px solid #d9ecff;
}
.form-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--el-color-primary);
  margin-bottom: 12px;
}
.form-row {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  color: #606266;
}
.form-label {
  color: #909399;
}
.form-upload-row {
  display: flex;
  align-items: center;
  margin-top: 12px;
}

.upload-list-preview { margin-top: 8px; }
.upload-item { display: flex; justify-content: space-between; padding: 4px 0; font-size: 13px; color: #606266; }
.upload-item-name { color: #409eff; }
.upload-item-size { color: #909399; }

.hidden-file-input {
  position: absolute;
  width: 0;
  height: 0;
  opacity: 0;
  pointer-events: none;
}
</style>

<style>
@media print {
  /* 隐藏侧边栏、顶栏 */
  .layout-aside,
  .layout-header { display: none !important; }
  /* 父容器取消固定高度，避免撑出空白页 */
  body, #app, .layout-container, .layout-main { height: auto !important; min-height: 0 !important; overflow: visible !important; }
  /* 隐藏订单列表搜索和表格区域 */
  .search-card,
  .table-card { display: none !important; }
  /* 去掉遮罩层背景，全部 static 布局 */
  .el-overlay { background: none !important; position: static !important; overflow: visible !important; }
  .el-overlay-dialog { position: static !important; height: auto !important; overflow: visible !important; }
  /* 对话框全宽展示 */
  .el-dialog { position: static !important; width: 100% !important; max-width: 100% !important; margin: 0 !important; box-shadow: none !important; }
  .el-dialog__header { display: none !important; }
  .el-dialog__body { padding: 20px !important; }
  .el-dialog__footer { display: none !important; }
  /* 隐藏操作类元素 */
  .inline-advance-form,
  .el-button { display: none !important; }
  .product-card { break-inside: avoid; }
}
</style>
