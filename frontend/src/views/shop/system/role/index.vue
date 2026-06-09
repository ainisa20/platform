<script setup lang="ts">
import { ref, reactive, onMounted, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import {
  getRoleList,
  getRoleById,
  createRole,
  updateRole,
  deleteRole,
  assignRolePermissions,
  getPermissionTree,
} from '@/api/shop/role'
import type {
  RoleResp,
  RoleCreateReq,
  PermissionResp,
} from '@/api/shop/role'

const DATA_SCOPE_MAP: Record<number, string> = {
  1: '全部数据',
  2: '本部门及下级',
}

const STATUS_MAP: Record<number, { label: string; type: 'success' | 'danger' }> = {
  1: { label: '启用', type: 'success' },
  2: { label: '禁用', type: 'danger' },
}

const loading = ref(false)
const tableData = ref<RoleResp[]>([])
const total = ref(0)

const queryParams = reactive({
  page: 1,
  page_size: 10,
  role_name: '',
  status: null as number | null,
})

async function fetchList() {
  loading.value = true
  try {
    const { data: res } = await getRoleList(queryParams)
    tableData.value = res.data.list
    total.value = res.data.total
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  queryParams.page = 1
  fetchList()
}

function handleReset() {
  queryParams.role_name = ''
  queryParams.status = null
  queryParams.page = 1
  fetchList()
}

function handleSizeChange(val: number) {
  queryParams.page_size = val
  queryParams.page = 1
  fetchList()
}

function handleCurrentChange(val: number) {
  queryParams.page = val
  fetchList()
}

const formDialogVisible = ref(false)
const formDialogTitle = ref('')
const formRef = ref<FormInstance>()
const editingId = ref<number | null>(null)

const formData = reactive<RoleCreateReq>({
  role_name: '',
  role_code: '',
  data_scope: 1,
  sort: 0,
  status: 1,
  remark: '',
})

const formRules = reactive<FormRules>({
  role_name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }],
  role_code: [{ required: true, message: '请输入角色编码', trigger: 'blur' }],
})

function openCreateDialog() {
  editingId.value = null
  formDialogTitle.value = '新增角色'
  resetFormData()
  formDialogVisible.value = true
  nextTick(() => formRef.value?.clearValidate())
}

function openEditDialog(row: RoleResp) {
  editingId.value = row.id
  formDialogTitle.value = '编辑角色'
  formData.role_name = row.role_name
  formData.role_code = row.role_code
  formData.data_scope = row.data_scope
  formData.sort = row.sort
  formData.status = row.status
  formData.remark = row.remark
  formDialogVisible.value = true
  nextTick(() => formRef.value?.clearValidate())
}

function resetFormData() {
  formData.role_name = ''
  formData.role_code = ''
  formData.data_scope = 1
  formData.sort = 0
  formData.status = 1
  formData.remark = ''
}

const formSubmitting = ref(false)

async function handleFormSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  formSubmitting.value = true
  try {
    if (editingId.value) {
      await updateRole(editingId.value, formData)
      ElMessage.success('更新成功')
    } else {
      await createRole(formData)
      ElMessage.success('创建成功')
    }
    formDialogVisible.value = false
    fetchList()
  } finally {
    formSubmitting.value = false
  }
}

async function handleDelete(row: RoleResp) {
  await ElMessageBox.confirm(`确定删除角色「${row.role_name}」吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
  })
  await deleteRole(row.id)
  ElMessage.success('删除成功')
  fetchList()
}

const permDialogVisible = ref(false)
const permTreeRef = ref<InstanceType<typeof import('element-plus')['ElTree']>>()
const permTreeData = ref<PermissionResp[]>([])
const permLoading = ref(false)
const currentPermRoleId = ref<number>(0)

function collectLeafIds(nodes: PermissionResp[]): number[] {
  const ids: number[] = []
  for (const node of nodes) {
    if (!node.children || node.children.length === 0) {
      ids.push(node.id)
    } else {
      ids.push(...collectLeafIds(node.children))
    }
  }
  return ids
}

async function openPermDialog(row: RoleResp) {
  currentPermRoleId.value = row.id
  permDialogVisible.value = true
  permLoading.value = true

  try {
    const { data: res } = await getPermissionTree()
    permTreeData.value = res.data

    const { data: detail } = await getRoleById(row.id)
    const rolePermIds = new Set(detail.data.permissions.map(p => p.id))
    const leafIds = collectLeafIds(res.data).filter(id => rolePermIds.has(id))

    permLoading.value = false
    await nextTick()
    permTreeRef.value?.setCheckedKeys(leafIds)
  } catch {
    permLoading.value = false
  }
}

const permSubmitting = ref(false)

async function handlePermSubmit() {
  permSubmitting.value = true
  try {
    const checkedKeys = permTreeRef.value?.getCheckedKeys() as number[]
    const halfCheckedKeys = permTreeRef.value?.getHalfCheckedKeys() as number[]
    const allIds = [...checkedKeys, ...halfCheckedKeys]
    await assignRolePermissions(currentPermRoleId.value, allIds)
    ElMessage.success('分配权限成功')
    permDialogVisible.value = false
    fetchList()
  } finally {
    permSubmitting.value = false
  }
}

onMounted(() => {
  fetchList()
})
</script>

<template>
  <div class="role-page">
    <div class="search-bar">
      <el-form :inline="true" :model="queryParams">
        <el-form-item label="角色名称">
          <el-input
            v-model="queryParams.role_name"
            placeholder="请输入角色名称"
            clearable
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="queryParams.status" placeholder="全部" clearable style="width: 120px">
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="action-bar">
      <el-button v-permission="'shop:role:create'" type="primary" @click="openCreateDialog">
        新增角色
      </el-button>
    </div>

    <el-table :data="tableData" v-loading="loading" border stripe style="width: 100%">
      <el-table-column prop="role_name" label="角色名称" min-width="120" />
      <el-table-column prop="role_code" label="角色编码" min-width="130" />
      <el-table-column prop="data_scope" label="数据范围" min-width="120">
        <template #default="{ row }">
          {{ DATA_SCOPE_MAP[row.data_scope] || '-' }}
        </template>
      </el-table-column>
      <el-table-column prop="sort" label="排序" width="80" align="center" />
      <el-table-column prop="status" label="状态" width="80" align="center">
        <template #default="{ row }">
          <el-tag :type="STATUS_MAP[row.status]?.type || 'info'" size="small">
            {{ STATUS_MAP[row.status]?.label || '未知' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="remark" label="备注" min-width="150" show-overflow-tooltip />
      <el-table-column prop="created_at" label="创建时间" min-width="170" />
      <el-table-column label="操作" width="220" fixed="right">
        <template #default="{ row }">
          <el-button v-permission="'shop:role:update'" type="primary" link size="small" @click="openEditDialog(row)">
            编辑
          </el-button>
          <el-button v-permission="'shop:role:assign'" type="primary" link size="small" @click="openPermDialog(row)">
            分配权限
          </el-button>
          <el-button v-permission="'shop:role:delete'" type="danger" link size="small" @click="handleDelete(row)">
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination-bar">
      <el-pagination
        v-model:current-page="queryParams.page"
        v-model:page-size="queryParams.page_size"
        :page-sizes="[10, 20, 50, 100]"
        :total="total"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>

    <el-dialog v-model="formDialogVisible" :title="formDialogTitle" width="520px" :close-on-click-modal="false" destroy-on-close>
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="90px">
        <el-form-item label="角色名称" prop="role_name">
          <el-input v-model="formData.role_name" placeholder="请输入角色名称" />
        </el-form-item>
        <el-form-item label="角色编码" prop="role_code">
          <el-input v-model="formData.role_code" placeholder="请输入角色编码" />
        </el-form-item>
        <el-form-item label="数据范围">
          <el-select v-model="formData.data_scope" style="width: 100%">
            <el-option :value="1" label="全部数据" />
            <el-option :value="2" label="本部门及下级" />
          </el-select>
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="formData.sort" :min="0" controls-position="right" style="width: 100%" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="formData.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="2">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="formData.remark" type="textarea" :rows="3" placeholder="请输入备注" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="formSubmitting" @click="handleFormSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="permDialogVisible" title="分配权限" width="560px" :close-on-click-modal="false" destroy-on-close>
      <el-tree
        v-if="!permLoading"
        ref="permTreeRef"
        :data="permTreeData"
        show-checkbox
        node-key="id"
        :props="{ label: 'name', children: 'children' }"
        default-expand-all
      >
        <template #default="{ data }">
          <span>{{ data.name }}<template v-if="data.perms_code">（{{ data.perms_code }}）</template></span>
        </template>
      </el-tree>
      <div v-else v-loading="true" style="height: 200px" />
      <template #footer>
        <el-button @click="permDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="permSubmitting" @click="handlePermSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.role-page {
  padding: 20px;
}

.search-bar {
  margin-bottom: 16px;
}

.action-bar {
  margin-bottom: 16px;
}

.pagination-bar {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>
