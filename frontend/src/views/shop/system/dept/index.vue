<script setup lang="ts">
import { ref, reactive, nextTick, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import {
  getDeptTree,
  createDept,
  updateDept,
  deleteDept,
  type DeptResp,
  type DeptCreateReq,
  type DeptUpdateReq,
} from '@/api/shop/dept'

const loading = ref(false)
const deptTree = ref<DeptResp[]>([])
const isExpandAll = ref(true)
const tableRef = ref()

const dialogVisible = ref(false)
const dialogTitle = ref('')
const dialogLoading = ref(false)
const isEdit = ref(false)
const editId = ref<number>(0)
const formRef = ref<FormInstance>()

const parentName = ref('')

interface DeptForm {
  parent_id: number
  dept_name: string
  leader: string
  phone: string
  sort: number
  status: number
}

const defaultForm = (): DeptForm => ({
  parent_id: 0,
  dept_name: '',
  leader: '',
  phone: '',
  sort: 0,
  status: 1,
})

const form = reactive<DeptForm>(defaultForm())

const rules = reactive<FormRules<DeptForm>>({
  dept_name: [{ required: true, message: '请输入部门名称', trigger: 'blur' }],
})

async function loadTree() {
  loading.value = true
  try {
    const { data: res } = await getDeptTree()
    deptTree.value = res.data || []
    await nextTick()
    if (isExpandAll.value) {
      toggleExpandAll()
    }
  } catch {
    // error already handled by request interceptor
  } finally {
    loading.value = false
  }
}

function toggleExpandAll() {
  isExpandAll.value = !isExpandAll.value
  const table = tableRef.value
  if (!table) return
  deptTree.value.forEach((row) => {
    table.toggleRowExpansion(row, isExpandAll.value)
    expandChildren(row, isExpandAll.value)
  })
}

function expandChildren(node: DeptResp, expanded: boolean) {
  const table = tableRef.value
  if (!table || !node.children?.length) return
  node.children.forEach((child) => {
    table.toggleRowExpansion(child, expanded)
    expandChildren(child, expanded)
  })
}

function handleCreate() {
  isEdit.value = false
  dialogTitle.value = '新增顶级部门'
  Object.assign(form, defaultForm())
  parentName.value = '无（顶级部门）'
  formRef.value?.resetFields()
  dialogVisible.value = true
}

function handleCreateChild(row: DeptResp) {
  isEdit.value = false
  dialogTitle.value = '新增子部门'
  Object.assign(form, defaultForm())
  form.parent_id = row.id
  parentName.value = row.dept_name
  formRef.value?.resetFields()
  dialogVisible.value = true
}

function handleEdit(row: DeptResp) {
  isEdit.value = true
  editId.value = row.id
  dialogTitle.value = '编辑部门'
  Object.assign(form, {
    parent_id: row.parent_id,
    dept_name: row.dept_name,
    leader: row.leader || '',
    phone: row.phone || '',
    sort: row.sort,
    status: row.status,
  })
  parentName.value = row.parent_id === 0 ? '无（顶级部门）' : findDeptName(deptTree.value, row.parent_id) || '未知'
  formRef.value?.resetFields()
  dialogVisible.value = true
}

function findDeptName(nodes: DeptResp[], id: number): string {
  for (const node of nodes) {
    if (node.id === id) return node.dept_name
    if (node.children?.length) {
      const found = findDeptName(node.children, id)
      if (found) return found
    }
  }
  return ''
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  dialogLoading.value = true
  try {
    if (isEdit.value) {
      const payload: DeptUpdateReq = {
        dept_name: form.dept_name,
        leader: form.leader || undefined,
        phone: form.phone || undefined,
        sort: form.sort,
        status: form.status,
      }
      await updateDept(editId.value, payload)
      ElMessage.success('更新成功')
    } else {
      const payload: DeptCreateReq = {
        parent_id: form.parent_id,
        dept_name: form.dept_name,
        leader: form.leader || undefined,
        phone: form.phone || undefined,
        sort: form.sort,
        status: form.status,
      }
      await createDept(payload)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    await loadTree()
  } catch {
    // error already handled by request interceptor
  } finally {
    dialogLoading.value = false
  }
}

async function handleDelete(row: DeptResp) {
  if (row.children?.length) {
    ElMessageBox.confirm(
      `部门「${row.dept_name}」下存在子部门，删除后将一并移除，确定要删除吗？`,
      '警告',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' },
    ).then(() => doDelete(row)).catch(() => {})
    return
  }
  try {
    await ElMessageBox.confirm(
      `确定要删除部门「${row.dept_name}」吗？`,
      '提示',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' },
    )
    await doDelete(row)
  } catch {
    // user cancelled
  }
}

async function doDelete(row: DeptResp) {
  try {
    await deleteDept(row.id)
    ElMessage.success('删除成功')
    await loadTree()
  } catch {
    // error already handled by request interceptor
  }
}

onMounted(() => {
  loadTree()
})
</script>

<template>
  <div class="dept-container">
    <div class="dept-header">
      <el-button v-permission="'shop:dept:create'" type="primary" @click="handleCreate">
        新增顶级部门
      </el-button>
      <el-button @click="toggleExpandAll">
        {{ isExpandAll ? '折叠全部' : '展开全部' }}
      </el-button>
    </div>

    <el-table
      ref="tableRef"
      v-loading="loading"
      :data="deptTree"
      row-key="id"
      :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
      border
      stripe
      default-expand-all
      class="dept-table"
    >
      <el-table-column prop="dept_name" label="部门名称" min-width="200" />
      <el-table-column prop="leader" label="负责人" width="120" />
      <el-table-column prop="phone" label="联系电话" width="140" />
      <el-table-column prop="sort" label="排序" width="80" align="center" />
      <el-table-column label="状态" width="90" align="center">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
            {{ row.status === 1 ? '启用' : '禁用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" width="180" />
      <el-table-column label="操作" width="220" fixed="right">
        <template #default="{ row }">
          <el-button
            v-permission="'shop:dept:create'"
            link
            type="primary"
            size="small"
            @click="handleCreateChild(row)"
          >
            新增子部门
          </el-button>
          <el-button
            v-permission="'shop:dept:update'"
            link
            type="primary"
            size="small"
            @click="handleEdit(row)"
          >
            编辑
          </el-button>
          <el-button
            v-permission="'shop:dept:delete'"
            link
            type="danger"
            size="small"
            @click="handleDelete(row)"
          >
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="520px"
      :close-on-click-modal="false" destroy-on-close
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="90px"
      >
        <el-form-item label="上级部门">
          <el-input :model-value="parentName" disabled />
        </el-form-item>
        <el-form-item label="部门名称" prop="dept_name">
          <el-input v-model="form.dept_name" placeholder="请输入部门名称" />
        </el-form-item>
        <el-form-item label="负责人" prop="leader">
          <el-input v-model="form.leader" placeholder="请输入负责人" />
        </el-form-item>
        <el-form-item label="联系电话" prop="phone">
          <el-input v-model="form.phone" placeholder="请输入联系电话" />
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="form.sort" :min="0" :max="9999" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="2">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="dialogLoading" @click="handleSubmit">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.dept-container {
  padding: 20px;
}

.dept-header {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}

.dept-table {
  width: 100%;
}
</style>
