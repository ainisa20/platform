<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import {
  getCategoryList,
  createCategory,
  updateCategory,
  deleteCategory,
} from '@/api/platform/product'
import type {
  ProductCategoryResp,
  ProductCategoryCreateReq,
  ProductCategoryUpdateReq,
} from '@/types/system'
import { initialsUpper } from '@/utils/code'

const loading = ref(false)
const tableData = ref<ProductCategoryResp[]>([])
const total = ref(0)

const searchForm = reactive({
  category_name: '',
})

const pagination = reactive({ page: 1, page_size: 10 })

async function fetchList() {
  loading.value = true
  try {
    const res = await getCategoryList({
      page: pagination.page,
      page_size: pagination.page_size,
      category_name: searchForm.category_name || undefined,
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
  searchForm.category_name = ''
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
const dialogTitle = ref('新增分类')
const isEdit = ref(false)
const editId = ref<number>(0)
const submitLoading = ref(false)
const formRef = ref<FormInstance>()

interface CategoryFormData {
  category_code: string
  category_name: string
  sort: number | undefined
  status: number
}

const defaultForm = (): CategoryFormData => ({
  category_code: '',
  category_name: '',
  sort: undefined,
  status: 1,
})

const formData = reactive<CategoryFormData>(defaultForm())

const formRules: FormRules = {
  category_name: [{ required: true, message: '请输入分类名称', trigger: 'blur' }],
  category_code: [{ required: true, message: '请输入分类编码', trigger: 'blur' }],
}

function autoFillCategoryCode() {
  if (isEdit.value) return
  formData.category_code = initialsUpper(formData.category_name)
}

function openCreateDialog() {
  isEdit.value = false
  dialogTitle.value = '新增分类'
  editId.value = 0
  Object.assign(formData, defaultForm())
  dialogVisible.value = true
}

function openEditDialog(row: ProductCategoryResp) {
  isEdit.value = true
  dialogTitle.value = '编辑分类'
  editId.value = row.id
  Object.assign(formData, {
    category_code: row.category_code,
    category_name: row.category_name,
    sort: row.sort,
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
      const updateData: ProductCategoryUpdateReq = {
        category_code: formData.category_code || undefined,
        category_name: formData.category_name,
        sort: formData.sort,
        status: formData.status,
      }
      await updateCategory(editId.value, updateData)
      ElMessage.success('更新成功')
    } else {
      const createData: ProductCategoryCreateReq = {
        category_code: formData.category_code || undefined,
        category_name: formData.category_name,
        sort: formData.sort,
        status: formData.status,
      }
      await createCategory(createData)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchList()
  } finally {
    submitLoading.value = false
  }
}

async function handleDelete(row: ProductCategoryResp) {
  await ElMessageBox.confirm(
    `确定要删除分类「${row.category_name}」吗？`,
    '删除确认',
    { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
  )
  await deleteCategory(row.id)
  ElMessage.success('删除成功')
  fetchList()
}

function formatTime(val: string | null): string {
  if (!val) return '-'
  return val.replace('T', ' ').substring(0, 19)
}

const statusTagType = (status: number) => (status === 1 ? 'success' : 'danger')
const statusText = (status: number) => (status === 1 ? '启用' : '停用')

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
        <el-button v-permission="'platform:product:category:create'" type="primary" @click="openCreateDialog">
          新增分类
        </el-button>
      </div>

      <el-table v-loading="loading" :data="tableData" border stripe style="width: 100%">
        <el-table-column prop="id" label="序号" width="70" />
        <el-table-column prop="category_code" label="分类编码" min-width="130" />
        <el-table-column prop="category_name" label="分类名称" min-width="140" />
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
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button v-permission="'platform:product:category:update'" type="primary" link size="small" @click="openEditDialog(row)">
              编辑
            </el-button>
            <el-button v-permission="'platform:product:category:delete'" type="danger" link size="small" @click="handleDelete(row)">
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

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="480px" :close-on-click-modal="false" destroy-on-close>
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="90px">
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
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="formData.sort" :min="0" style="width: 100%" />
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
  </div>
</template>

<style scoped>
.category-page { padding: 16px; }
.search-card { margin-bottom: 16px; }
.search-card :deep(.el-card__body) { padding-bottom: 2px; }
.table-header { margin-bottom: 16px; }
.pagination-wrap { display: flex; justify-content: flex-end; margin-top: 16px; }
</style>
