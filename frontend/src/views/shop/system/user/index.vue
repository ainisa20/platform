<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import {
  getUserList,
  createUser,
  updateUser,
  deleteUser,
  assignUserRoles,
  resetUserPassword,
  getAssignableRoles,
  getDeptTree,
} from '@/api/shop/user'
import type {
  UserResp,
  UserCreateReq,
  UserUpdateReq,
  DeptResp,
  RoleResp,
} from '@/types/system'

// ==================== 列表数据 ====================

const loading = ref(false)
const tableData = ref<UserResp[]>([])
const total = ref(0)

const searchForm = reactive({
  username: '',
  real_name: '',
  phone: '',
  status: null as number | null,
})

const pagination = reactive({
  page: 1,
  page_size: 10,
})

async function fetchList() {
  loading.value = true
  try {
    const res = await getUserList({
      page: pagination.page,
      page_size: pagination.page_size,
      username: searchForm.username || undefined,
      real_name: searchForm.real_name || undefined,
      phone: searchForm.phone || undefined,
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
  searchForm.username = ''
  searchForm.real_name = ''
  searchForm.phone = ''
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

// ==================== 角色与部门数据 ====================

const roleOptions = ref<RoleResp[]>([])
const deptTreeData = ref<DeptResp[]>([])
const deptMap = computed(() => {
  const map = new Map<number, DeptResp>()
  const walk = (nodes: DeptResp[]) => {
    for (const n of nodes) {
      map.set(n.id, n)
      if (n.children?.length) walk(n.children)
    }
  }
  walk(deptTreeData.value)
  return map
})

async function loadDropdownData() {
  const [rolesRes, deptsRes] = await Promise.all([
    getAssignableRoles(),
    getDeptTree(),
  ])
  roleOptions.value = rolesRes.data.data
  deptTreeData.value = deptsRes.data.data
}

// ==================== 新增/编辑弹窗 ====================

const dialogVisible = ref(false)
const dialogTitle = ref('新增用户')
const isEdit = ref(false)
const editUserId = ref<number>(0)
const formRef = ref<FormInstance>()
const submitLoading = ref(false)

const defaultForm = (): UserCreateReq => ({
  username: '',
  password: '',
  real_name: '',
  phone: '',
  email: '',
  dept_id: null,
  role_ids: [],
  status: 1,
})

const formData = reactive<UserCreateReq>(defaultForm())

const formRules = computed<FormRules>(() => ({
  username: isEdit.value
    ? []
    : [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: isEdit.value
    ? []
    : [
        { required: true, message: '请输入密码', trigger: 'blur' },
        { min: 6, message: '密码长度不能少于6位', trigger: 'blur' },
      ],
  real_name: [{ required: true, message: '请输入姓名', trigger: 'blur' }],
  email: [{ type: 'email', message: '请输入正确的邮箱地址', trigger: 'blur' }],
  dept_id: isEdit.value
    ? []
    : [{ required: true, message: '请选择部门', trigger: 'change' }],
  role_ids: isEdit.value
    ? []
    : [{ required: true, message: '请选择角色', trigger: 'change' }],
}))

function openCreateDialog() {
  isEdit.value = false
  dialogTitle.value = '新增用户'
  editUserId.value = 0
  Object.assign(formData, defaultForm())
  dialogVisible.value = true
}

function openEditDialog(row: UserResp) {
  isEdit.value = true
  dialogTitle.value = '编辑用户'
  editUserId.value = row.id
  Object.assign(formData, {
    username: row.username,
    password: '',
    real_name: row.real_name,
    phone: row.phone,
    email: row.email,
    dept_id: row.dept_id,
    role_ids: row.roles.map((r) => r.id),
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
      const updateData: UserUpdateReq = {
        real_name: formData.real_name,
        phone: formData.phone || undefined,
        email: formData.email || undefined,
        dept_id: formData.dept_id,
        role_ids: formData.role_ids,
        status: formData.status,
      }
      await updateUser(editUserId.value, updateData)
      ElMessage.success('更新成功')
    } else {
      await createUser(formData)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchList()
  } finally {
    submitLoading.value = false
  }
}

// ==================== 删除 ====================

async function handleDelete(row: UserResp) {
  await ElMessageBox.confirm(`确定要删除用户「${row.real_name || row.username}」吗？`, '删除确认', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
  })
  await deleteUser(row.id)
  ElMessage.success('删除成功')
  fetchList()
}

// ==================== 分配角色弹窗 ====================

const assignDialogVisible = ref(false)
const assignUserId = ref(0)
const assignRoleIds = ref<number[]>([])
const assignLoading = ref(false)

function openAssignDialog(row: UserResp) {
  assignUserId.value = row.id
  assignRoleIds.value = row.roles.map((r) => r.id)
  assignDialogVisible.value = true
}

async function handleAssignSubmit() {
  assignLoading.value = true
  try {
    await assignUserRoles(assignUserId.value, { role_ids: assignRoleIds.value })
    ElMessage.success('分配成功')
    assignDialogVisible.value = false
    fetchList()
  } finally {
    assignLoading.value = false
  }
}

// ==================== 重置密码弹窗 ====================

const resetDialogVisible = ref(false)
const resetUserId = ref(0)
const resetNewPassword = ref('')
const resetLoading = ref(false)
const resetFormRef = ref<FormInstance>()

const resetRules: FormRules = {
  new_password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' },
  ],
}

function openResetDialog(row: UserResp) {
  resetUserId.value = row.id
  resetNewPassword.value = ''
  resetDialogVisible.value = true
}

async function handleResetSubmit() {
  const valid = await resetFormRef.value?.validate().catch(() => false)
  if (!valid) return

  resetLoading.value = true
  try {
    await resetUserPassword(resetUserId.value, { new_password: resetNewPassword.value })
    ElMessage.success('密码重置成功')
    resetDialogVisible.value = false
  } finally {
    resetLoading.value = false
  }
}

// ==================== 辅助函数 ====================

function formatTime(val: string | null): string {
  if (!val) return '-'
  return val.replace('T', ' ').substring(0, 19)
}

const statusTagType = (status: number) => (status === 1 ? 'success' : 'danger')
const statusText = (status: number) => (status === 1 ? '启用' : '禁用')

const deptTreeProps = {
  value: 'id',
  label: 'dept_name',
  children: 'children',
}

// ==================== 初始化 ====================

onMounted(() => {
  fetchList()
  loadDropdownData()
})
</script>

<template>
  <div class="user-page">
    <!-- 搜索栏 -->
    <el-card shadow="never" class="search-card">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="用户名">
          <el-input v-model="searchForm.username" placeholder="请输入用户名" clearable />
        </el-form-item>
        <el-form-item label="姓名">
          <el-input v-model="searchForm.real_name" placeholder="请输入姓名" clearable />
        </el-form-item>
        <el-form-item label="手机号">
          <el-input v-model="searchForm.phone" placeholder="请输入手机号" clearable />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部" clearable style="width: 120px">
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 表格 -->
    <el-card shadow="never" class="table-card">
      <div class="table-header">
        <el-button v-permission="'shop:user:create'" type="primary" @click="openCreateDialog">
          新增用户
        </el-button>
      </div>

      <el-table v-loading="loading" :data="tableData" border stripe style="width: 100%">
        <el-table-column prop="username" label="用户名" min-width="120" />
        <el-table-column prop="real_name" label="姓名" min-width="100" />
        <el-table-column prop="phone" label="手机号" min-width="130" />
        <el-table-column label="部门" min-width="120">
          <template #default="{ row }">
            {{ deptMap.get(row.dept_id)?.dept_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="角色" min-width="160">
          <template #default="{ row }">
            <template v-if="row.roles?.length">
              <el-tag v-for="role in row.roles" :key="role.id" size="small" class="role-tag">
                {{ role.role_name }}
              </el-tag>
            </template>
            <span v-else>-</span>
          </template>
        </el-table-column>
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
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="{ row }">
            <el-button v-permission="'shop:user:update'" type="primary" link size="small" @click="openEditDialog(row)">
              编辑
            </el-button>
            <el-button v-permission="'shop:user:assign'" type="warning" link size="small" @click="openAssignDialog(row)">
              分配角色
            </el-button>
            <el-button v-permission="'shop:user:reset'" type="info" link size="small" @click="openResetDialog(row)">
              重置密码
            </el-button>
            <el-button v-permission="'shop:user:delete'" type="danger" link size="small" @click="handleDelete(row)">
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

    <!-- 新增/编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="560px" :close-on-click-modal="false" destroy-on-close>
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="80px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="formData.username" placeholder="请输入用户名" :disabled="isEdit" />
        </el-form-item>
        <el-form-item v-if="!isEdit" label="密码" prop="password">
          <el-input v-model="formData.password" type="password" placeholder="请输入密码（至少6位）" show-password />
        </el-form-item>
        <el-form-item label="姓名" prop="real_name">
          <el-input v-model="formData.real_name" placeholder="请输入姓名" />
        </el-form-item>
        <el-form-item label="手机号" prop="phone">
          <el-input v-model="formData.phone" placeholder="请输入手机号" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="formData.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="部门" prop="dept_id">
          <el-tree-select
            v-model="formData.dept_id"
            :data="deptTreeData"
            :props="deptTreeProps"
            placeholder="请选择部门"
            check-strictly
            clearable
            filterable
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="角色" prop="role_ids">
          <el-select v-model="formData.role_ids" multiple placeholder="请选择角色" style="width: 100%">
            <el-option v-for="role in roleOptions" :key="role.id" :label="role.role_name" :value="role.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="formData.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="2">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 分配角色弹窗 -->
    <el-dialog v-model="assignDialogVisible" title="分配角色" width="440px" :close-on-click-modal="false" destroy-on-close>
      <el-checkbox-group v-model="assignRoleIds">
        <el-checkbox v-for="role in roleOptions" :key="role.id" :value="role.id" :label="role.role_name" />
      </el-checkbox-group>
      <template #footer>
        <el-button @click="assignDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="assignLoading" @click="handleAssignSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 重置密码弹窗 -->
    <el-dialog v-model="resetDialogVisible" title="重置密码" width="420px" :close-on-click-modal="false" destroy-on-close>
      <el-form ref="resetFormRef" :model="{ new_password: resetNewPassword }" :rules="resetRules" label-width="80px">
        <el-form-item label="新密码" prop="new_password">
          <el-input v-model="resetNewPassword" type="password" placeholder="请输入新密码（至少6位）" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="resetDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="resetLoading" @click="handleResetSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.user-page {
  padding: 16px;
}

.search-card {
  margin-bottom: 16px;
}

.search-card :deep(.el-card__body) {
  padding-bottom: 2px;
}

.table-header {
  margin-bottom: 16px;
}

.role-tag {
  margin: 2px 4px 2px 0;
}

.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
</style>
