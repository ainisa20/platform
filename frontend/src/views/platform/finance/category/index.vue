<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import {
  getFinanceCategoryList,
  createFinanceCategory,
  updateFinanceCategory,
  deleteFinanceCategory,
} from '@/api/platform/finance'
import type {
  FinanceCategoryResp,
  FinanceCategoryCreateReq,
  FinanceCategoryUpdateReq,
} from '@/types/system'
import { initialsUpper } from '@/utils/code'

const loading = ref(false)
const tableData = ref<FinanceCategoryResp[]>([])
const activeType = ref<number | null>(null)

const searchForm = reactive({
  category_name: '',
})

async function fetchList() {
  loading.value = true
  try {
    const res = await getFinanceCategoryList({
      category_type: activeType.value,
      category_name: searchForm.category_name || undefined,
    })
    tableData.value = res.data.data
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  fetchList()
}

function handleReset() {
  searchForm.category_name = ''
  fetchList()
}

function handleTypeChange() {
  fetchList()
}

const dialogVisible = ref(false)
const dialogTitle = ref('新增分类')
const isEdit = ref(false)
const editId = ref<number>(0)
const submitLoading = ref(false)
const formRef = ref<FormInstance>()

interface CategoryFormData {
  parent_id: number
  category_type: number
  category_code: string
  category_name: string
  finance_code: string
  sort: number | undefined
}

const defaultForm = (): CategoryFormData => ({
  parent_id: 0,
  category_type: 1,
  category_code: '',
  category_name: '',
  finance_code: '',
  sort: undefined,
})

const formData = reactive<CategoryFormData>(defaultForm())

const formRules: FormRules = {
  category_name: [{ required: true, message: '请输入分类名称', trigger: 'blur' }],
}

const isParentTypeLocked = computed(() => {
  return formData.parent_id !== 0
})

const parentTypeDisabled = computed(() => {
  return isEdit.value || isParentTypeLocked.value
})

const treeSelectData = computed(() => {
  return filterTreeForSelect(tableData.value, editId.value)
})

function filterTreeForSelect(nodes: FinanceCategoryResp[], excludeId: number): FinanceCategoryResp[] {
  const result: FinanceCategoryResp[] = []
  for (const node of nodes) {
    if (node.id === excludeId) continue
    if (node.level >= 3) {
      result.push({ ...node, children: [] })
    } else {
      const filtered = node.children ? filterTreeForSelect(node.children, excludeId) : []
      result.push({ ...node, children: filtered.length ? filtered : undefined })
    }
  }
  return result
}

function findNodeById(nodes: FinanceCategoryResp[], id: number): FinanceCategoryResp | null {
  for (const node of nodes) {
    if (node.id === id) return node
    if (node.children) {
      const found = findNodeById(node.children, id)
      if (found) return found
    }
  }
  return null
}

function autoFillCategoryCode() {
  if (isEdit.value) return
  formData.category_code = initialsUpper(formData.category_name)
}

function handleParentChange(parentId: number) {
  if (parentId === 0) {
    formData.category_type = 1
  } else {
    const parentNode = findNodeById(tableData.value, parentId)
    if (parentNode) {
      formData.category_type = parentNode.category_type
    }
  }
}

function openCreateDialog(parentRow?: FinanceCategoryResp) {
  isEdit.value = false
  dialogTitle.value = '新增分类'
  editId.value = 0
  Object.assign(formData, defaultForm())
  if (parentRow) {
    formData.parent_id = parentRow.id
    formData.category_type = parentRow.category_type
  }
  dialogVisible.value = true
}

function openEditDialog(row: FinanceCategoryResp) {
  isEdit.value = true
  dialogTitle.value = '编辑分类'
  editId.value = row.id
  Object.assign(formData, {
    parent_id: row.parent_id,
    category_type: row.category_type,
    category_code: row.category_code,
    category_name: row.category_name,
    finance_code: row.finance_code,
    sort: row.sort,
  })
  dialogVisible.value = true
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    if (isEdit.value) {
      const updateData: FinanceCategoryUpdateReq = {
        parent_id: formData.parent_id || undefined,
        category_type: formData.category_type,
        category_code: formData.category_code || undefined,
        category_name: formData.category_name,
        finance_code: formData.finance_code || undefined,
        sort: formData.sort,
      }
      await updateFinanceCategory(editId.value, updateData)
      ElMessage.success('更新成功')
    } else {
      const createData: FinanceCategoryCreateReq = {
        parent_id: formData.parent_id || undefined,
        category_type: formData.category_type,
        category_code: formData.category_code || undefined,
        category_name: formData.category_name,
        finance_code: formData.finance_code || undefined,
        sort: formData.sort,
      }
      await createFinanceCategory(createData)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchList()
  } finally {
    submitLoading.value = false
  }
}

async function handleDelete(row: FinanceCategoryResp) {
  await ElMessageBox.confirm(
    `确定要删除分类「${row.category_name}」吗？`,
    '删除确认',
    { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
  )
  await deleteFinanceCategory(row.id)
  ElMessage.success('删除成功')
  fetchList()
}

function formatTime(val: string | null): string {
  if (!val) return '-'
  return val.replace('T', ' ').substring(0, 19)
}

const typeTagType = (type: number) => (type === 1 ? 'success' : 'warning')
const typeText = (type: number) => (type === 1 ? '收入' : '支出')

onMounted(fetchList)
</script>

<template>
  <div class="category-page">
    <el-card shadow="never" class="search-card">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="分类名称">
          <el-input v-model="searchForm.category_name" placeholder="请输入分类名称" clearable />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never" class="table-card">
      <div class="table-header">
        <el-radio-group v-model="activeType" @change="handleTypeChange" style="margin-right: 16px">
          <el-radio-button :value="null">全部</el-radio-button>
          <el-radio-button :value="1">收入</el-radio-button>
          <el-radio-button :value="2">支出</el-radio-button>
        </el-radio-group>
        <el-button v-permission="'platform:finance:category:create'" type="primary" @click="openCreateDialog()">
          新增一级分类
        </el-button>
      </div>

      <el-table
        v-loading="loading"
        :data="tableData"
        border
        stripe
        row-key="id"
        :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
        style="width: 100%"
      >
        <el-table-column prop="category_code" label="分类编码" min-width="130" />
        <el-table-column prop="category_name" label="分类名称" min-width="140" />
        <el-table-column label="类型" min-width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="typeTagType(row.category_type)" size="small">
              {{ typeText(row.category_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="finance_code" label="财务编号" min-width="120" />
        <el-table-column prop="sort" label="排序" width="80" align="center" />
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
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button v-permission="'platform:finance:category:update'" type="primary" link size="small" @click="openEditDialog(row)">
              编辑
            </el-button>
            <el-button
              v-if="row.level < 3"
              v-permission="'platform:finance:category:create'"
              type="primary"
              link
              size="small"
              @click="openCreateDialog(row)"
            >
              新增子分类
            </el-button>
            <el-button v-permission="'platform:finance:category:delete'" type="danger" link size="small" @click="handleDelete(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="520px" :close-on-click-modal="false" destroy-on-close>
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="90px">
        <el-form-item v-if="formData.parent_id !== 0 || isEdit" label="上级分类" prop="parent_id">
          <el-tree-select
            v-model="formData.parent_id"
            :data="treeSelectData"
            :props="{ label: 'category_name', value: 'id', children: 'children' }"
            placeholder="请选择上级分类"
            check-strictly
            disabled
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="分类类型" prop="category_type">
          <el-radio-group v-model="formData.category_type" :disabled="parentTypeDisabled">
            <el-radio :value="1">收入</el-radio>
            <el-radio :value="2">支出</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="分类名称" prop="category_name">
          <el-input
            v-model="formData.category_name"
            placeholder="请输入分类名称"
            @input="autoFillCategoryCode"
          />
        </el-form-item>
        <el-form-item label="分类编码" prop="category_code">
          <el-input v-model="formData.category_code" placeholder="输入名称后自动生成，可手动修改" />
        </el-form-item>
        <el-form-item label="财务编号" prop="finance_code">
          <el-input v-model="formData.finance_code" placeholder="请输入财务编号" />
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="formData.sort" :min="0" style="width: 100%" />
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
.category-page { padding: 16px; }
.search-card { margin-bottom: 16px; }
.search-card :deep(.el-card__body) { padding-bottom: 2px; }
.table-header { display: flex; align-items: center; margin-bottom: 16px; }
</style>
