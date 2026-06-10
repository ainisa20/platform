<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import {
  getProductList,
  getProduct,
  createProduct,
  updateProduct,
  deleteProduct,
  updateProductStatus,
  getCategoryList,
  getProductWorkflow,
  saveProductWorkflow,
} from '@/api/platform/product'
import type {
  ProductResp,
  ProductCreateReq,
  ProductUpdateReq,
  ProductCategoryResp,
  WorkflowNodeReq,
} from '@/types/system'
import { pinyin } from 'pinyin-pro'
import { initialsUpper } from '@/utils/code'

const loading = ref(false)
const tableData = ref<ProductResp[]>([])
const total = ref(0)

const searchForm = reactive({
  product_name: '',
  category_id: null as number | null,
  status: null as number | null,
})

const pagination = reactive({ page: 1, page_size: 10 })

async function fetchList() {
  loading.value = true
  try {
    const res = await getProductList({
      page: pagination.page,
      page_size: pagination.page_size,
      product_name: searchForm.product_name || undefined,
      category_id: searchForm.category_id ?? undefined,
      status: searchForm.status ?? undefined,
    })
    tableData.value = res.data.data.list
    total.value = res.data.data.total
  } catch {
    tableData.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  pagination.page = 1
  fetchList()
}

function handleReset() {
  searchForm.product_name = ''
  searchForm.category_id = null
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

const categoryOptions = ref<ProductCategoryResp[]>([])

async function loadCategories() {
  try {
    const res = await getCategoryList({ page: 1, page_size: 100 })
    categoryOptions.value = res.data.data.list || []
  } catch {
    categoryOptions.value = []
  }
}

const dialogVisible = ref(false)
const dialogTitle = ref('新增商品')
const isEdit = ref(false)
const editId = ref<number>(0)
const submitLoading = ref(false)
const formRef = ref<FormInstance>()

interface ProductFormData {
  product_code: string
  product_name: string
  category_id: number | undefined
  price: number | undefined
  sort: number | undefined
  status: number
  description: string
}

const defaultForm = (): ProductFormData => ({
  product_code: '',
  product_name: '',
  category_id: undefined,
  price: undefined,
  sort: undefined,
  status: 1,
  description: '',
})

const formData = reactive<ProductFormData>(defaultForm())

const formRules = reactive<FormRules>({
  product_code: [{ required: true, message: '请输入商品编号', trigger: 'blur' }],
  product_name: [{ required: true, message: '请输入商品名称', trigger: 'blur' }],
  category_id: [{ required: true, message: '请选择商品分类', trigger: 'change' }],
  price: [{ required: true, message: '请输入标准价格', trigger: 'blur' }],
})

function openCreateDialog() {
  isEdit.value = false
  dialogTitle.value = '新增商品'
  editId.value = 0
  Object.assign(formData, defaultForm())
  dialogVisible.value = true
}

function autoFillProductCode() {
  if (isEdit.value) return
  formData.product_code = initialsUpper(formData.product_name)
}

function openEditDialog(row: ProductResp) {
  isEdit.value = true
  dialogTitle.value = '编辑商品'
  editId.value = row.id
  Object.assign(formData, {
    product_code: row.product_code,
    product_name: row.product_name,
    category_id: row.category_id,
    price: row.price,
    sort: row.sort,
    status: row.status,
    description: row.description,
  })
  dialogVisible.value = true
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    if (isEdit.value) {
      const updateData: ProductUpdateReq = {
        product_name: formData.product_name,
        category_id: formData.category_id,
        price: formData.price,
        sort: formData.sort,
        status: formData.status,
        description: formData.description || undefined,
      }
      await updateProduct(editId.value, updateData)
      ElMessage.success('更新成功')
    } else {
      const createData: ProductCreateReq = {
        product_code: formData.product_code,
        product_name: formData.product_name,
        category_id: formData.category_id,
        price: formData.price ?? 0,
        sort: formData.sort,
        status: formData.status,
        description: formData.description || undefined,
      }
      await createProduct(createData)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchList()
  } finally {
    submitLoading.value = false
  }
}

async function handleDelete(row: ProductResp) {
  await ElMessageBox.confirm(
    `确定要删除商品「${row.product_name}」吗？`,
    '删除确认',
    { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
  )
  await deleteProduct(row.id)
  ElMessage.success('删除成功')
  fetchList()
}

async function handleToggleStatus(row: ProductResp) {
  const next = row.status === 1 ? 2 : 1
  const action = next === 1 ? '上架' : '下架'
  await ElMessageBox.confirm(
    `确定要${action}商品「${row.product_name}」吗？`,
    `${action}确认`,
    { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
  )
  await updateProductStatus(row.id, { status: next })
  ElMessage.success(`${action}成功`)
  fetchList()
}

const workflowDialogVisible = ref(false)
const workflowProductId = ref(0)
const workflowProductName = ref('')
const workflowProductCode = ref('')
const workflowLoading = ref(false)
const workflowNodes = ref<WorkflowNodeReq[]>([])

function buildNodeCode(productCode: string, nodeName: string): string {
  if (!nodeName) return ''
  const initials = pinyin(nodeName, { pattern: 'first', toneType: 'none', type: 'array' })
    .join('')
    .replace(/[^a-zA-Z0-9]/g, '')
    .toLowerCase()
  return `${productCode}_${initials}`
}

const defaultWorkflowNames = ['联系客户', '下单', '创建账号', '部署', '实施', '完成']

function buildDefaultNodes(productCode: string): WorkflowNodeReq[] {
  return defaultWorkflowNames.map((name, i) => ({
    node_index: i + 1,
    node_code: buildNodeCode(productCode, name),
    node_name: name,
  }))
}

function autoFillNodeCode(node: WorkflowNodeReq) {
  if (!workflowProductCode.value) return
  node.node_code = buildNodeCode(workflowProductCode.value, node.node_name)
}

async function openWorkflowDialog(row: ProductResp) {
  workflowProductId.value = row.id
  workflowProductName.value = row.product_name
  workflowProductCode.value = row.product_code
  workflowDialogVisible.value = true
  workflowLoading.value = true
  try {
    const res = await getProductWorkflow(row.id)
    const nodes = res.data.data
    if (nodes && nodes.length > 0) {
      workflowNodes.value = nodes.map((n) => ({
        node_index: n.node_index,
        node_code: n.node_code,
        node_name: n.node_name,
      }))
    } else {
      workflowNodes.value = buildDefaultNodes(row.product_code)
    }
  } finally {
    workflowLoading.value = false
  }
}

function addWorkflowNode() {
  const nextIndex = workflowNodes.value.length + 1
  const newNode: WorkflowNodeReq = { node_index: nextIndex, node_code: '', node_name: '' }
  workflowNodes.value.push(newNode)
}

function removeWorkflowNode(index: number) {
  workflowNodes.value.splice(index, 1)
  workflowNodes.value.forEach((node, i) => {
    node.node_index = i + 1
  })
}

async function handleSaveWorkflow() {
  for (const node of workflowNodes.value) {
    if (!node.node_code || !node.node_name) {
      ElMessage.warning('请填写所有节点的编码和名称')
      return
    }
  }
  workflowLoading.value = true
  try {
    await saveProductWorkflow(workflowProductId.value, {
      nodes: workflowNodes.value,
    })
    ElMessage.success('流程配置保存成功')
    workflowDialogVisible.value = false
  } finally {
    workflowLoading.value = false
  }
}

function formatTime(val: string | null): string {
  if (!val) return '-'
  return val.replace('T', ' ').substring(0, 19)
}

function formatPrice(val: number): string {
  return Number(val).toFixed(2)
}

const statusTagType = (status: number) => (status === 1 ? 'success' : 'danger')
const statusText = (status: number) => (status === 1 ? '上架' : '下架')

onMounted(() => {
  fetchList()
  loadCategories()
})
</script>

<template>
  <div class="product-page">
    <el-card shadow="never" class="search-card">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="商品名称">
          <el-input v-model="searchForm.product_name" placeholder="请输入商品名称" clearable />
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="searchForm.category_id" placeholder="全部分类" clearable style="width: 160px">
            <el-option
              v-for="cat in categoryOptions"
              :key="cat.id"
              :label="cat.category_name"
              :value="cat.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部状态" clearable style="width: 120px">
            <el-option label="上架" :value="1" />
            <el-option label="下架" :value="2" />
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
        <el-button v-permission="'platform:product:create'" type="primary" @click="openCreateDialog">
          新增商品
        </el-button>
      </div>

      <el-table v-loading="loading" :data="tableData" border stripe style="width: 100%">
        <el-table-column prop="product_code" label="商品编号" min-width="130" />
        <el-table-column prop="product_name" label="商品名称" min-width="140" />
        <el-table-column label="分类" min-width="120">
          <template #default="{ row }">
            {{ row.category_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="标准价格" min-width="110" align="right">
          <template #default="{ row }">
            {{ formatPrice(row.price) }}
          </template>
        </el-table-column>
        <el-table-column prop="sort" label="排序" width="80" align="center" />
        <el-table-column label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">
              {{ statusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" min-width="170">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="创建人" min-width="100">
          <template #default="{ row }">
            {{ row.created_by || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button v-permission="'platform:product:update'" type="primary" link size="small" @click="openEditDialog(row)">
              编辑
            </el-button>
            <el-button
              v-permission="'platform:product:status'"
              :type="row.status === 1 ? 'warning' : 'success'" link size="small"
              @click="handleToggleStatus(row)"
            >
              {{ row.status === 1 ? '下架' : '上架' }}
            </el-button>
            <el-button v-permission="'platform:product:update'" type="info" link size="small" @click="openWorkflowDialog(row)">
              流程配置
            </el-button>
            <el-button v-permission="'platform:product:delete'" type="danger" link size="small" @click="handleDelete(row)">
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

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="600px" :close-on-click-modal="false" destroy-on-close>
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="90px">
        <el-form-item label="商品名称" prop="product_name">
          <el-input
            v-model="formData.product_name"
            placeholder="请输入商品名称"
            @input="autoFillProductCode"
          />
        </el-form-item>
        <el-form-item label="商品编号" prop="product_code">
          <el-input
            v-model="formData.product_code"
            placeholder="输入名称后自动生成，可手动修改"
            :disabled="isEdit"
          />
        </el-form-item>
        <el-form-item label="分类" prop="category_id">
          <el-select v-model="formData.category_id" placeholder="请选择分类" style="width: 100%">
            <el-option
              v-for="cat in categoryOptions"
              :key="cat.id"
              :label="cat.category_name"
              :value="cat.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="标准价格" prop="price">
          <el-input-number v-model="formData.price" :precision="2" :min="0" :step="100" style="width: 100%" />
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="formData.sort" :min="0" style="width: 100%" />
        </el-form-item>
        <el-form-item label="商品描述" prop="description">
          <el-input v-model="formData.description" type="textarea" :rows="3" placeholder="请输入商品描述" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="workflowDialogVisible" :title="`流程配置 - ${workflowProductName}`" width="640px" :close-on-click-modal="false" destroy-on-close>
      <div v-loading="workflowLoading">
        <div class="workflow-header">
          <el-button type="primary" size="small" @click="addWorkflowNode">添加节点</el-button>
        </div>
        <el-table :data="workflowNodes" border style="width: 100%">
          <el-table-column label="序号" width="70" align="center">
            <template #default="{ row }">
              {{ row.node_index }}
            </template>
          </el-table-column>
          <el-table-column label="节点名称" min-width="180">
            <template #default="{ row }">
              <el-input
                v-model="row.node_name"
                placeholder="请输入节点名称"
                size="small"
                @input="autoFillNodeCode(row)"
              />
            </template>
          </el-table-column>
          <el-table-column label="节点编码" min-width="200">
            <template #default="{ row }">
              <el-input v-model="row.node_code" placeholder="输入节点名称后自动生成，可手动修改" size="small" />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="80" align="center">
            <template #default="{ $index }">
              <el-button type="danger" link size="small" @click="removeWorkflowNode($index)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
      <template #footer>
        <el-button @click="workflowDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="workflowLoading" @click="handleSaveWorkflow">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.product-page { padding: 16px; }
.search-card { margin-bottom: 16px; }
.search-card :deep(.el-card__body) { padding-bottom: 2px; }
.table-header { margin-bottom: 16px; }
.pagination-wrap { display: flex; justify-content: flex-end; margin-top: 16px; }
.workflow-header { margin-bottom: 12px; }
</style>
