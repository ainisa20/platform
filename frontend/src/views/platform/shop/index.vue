<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { regionData, codeToText } from 'element-china-area-data'
import {
  getShopList,
  getShop,
  createShop,
  updateShop,
  deleteShop,
  updateShopStatus,
  resetShopAdminPassword,
} from '@/api/platform/shop'
import type {
  ShopResp,
  ShopCreateReq,
  ShopUpdateReq,
} from '@/types/system'

/** Find cascader value array from region text names */
function findRegionCodes(province: string, city: string, district: string): string[] {
  for (const p of regionData) {
    if (p.label === province) {
      if (!city) return [p.value]
      for (const c of p.children ?? []) {
        if (c.label === city) {
          if (!district) return [p.value, c.value]
          for (const d of c.children ?? []) {
            if (d.label === district) {
              return [p.value, c.value, d.value]
            }
          }
          return [p.value, c.value]
        }
      }
      return [p.value]
    }
  }
  return []
}

const loading = ref(false)
const tableData = ref<ShopResp[]>([])
const total = ref(0)

const searchForm = reactive({
  shop_code: '',
  shop_name: '',
  province: '',
  city: '',
  district: '',
  status: null as number | null,
})

/** Selected region codes for search cascader */
const searchRegion = ref<string[]>([])

const pagination = reactive({ page: 1, page_size: 10 })

async function fetchList() {
  loading.value = true
  try {
    // Convert search region codes to text names
    if (searchRegion.value.length > 0) {
      searchForm.province = codeToText[searchRegion.value[0]] ?? ''
      searchForm.city = searchRegion.value[1] ? (codeToText[searchRegion.value[1]] ?? '') : ''
      searchForm.district = searchRegion.value[2] ? (codeToText[searchRegion.value[2]] ?? '') : ''
    } else {
      searchForm.province = ''
      searchForm.city = ''
      searchForm.district = ''
    }
    const res = await getShopList({
      page: pagination.page,
      page_size: pagination.page_size,
      shop_code: searchForm.shop_code || undefined,
      shop_name: searchForm.shop_name || undefined,
      province: searchForm.province || undefined,
      city: searchForm.city || undefined,
      district: searchForm.district || undefined,
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
  searchForm.shop_code = ''
  searchForm.shop_name = ''
  searchForm.province = ''
  searchForm.city = ''
  searchForm.district = ''
  searchForm.status = null
  searchRegion.value = []
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
const dialogTitle = ref('新增店铺')
const isEdit = ref(false)
const editId = ref<number>(0)
const submitLoading = ref(false)
const formRef = ref<FormInstance>()

/** Selected region codes for cascader: [province, city, district] */
const selectedRegion = ref<string[]>([])

const createForm = reactive<ShopCreateReq>({
  shop_code: '',
  shop_name: '',
  contact: '',
  phone: '',
  email: '',
  province: '',
  city: '',
  district: '',
  detail_address: '',
  address: '',
  remark: '',
  admin_username: '',
  admin_password: '',
  admin_real_name: '',
})

const editForm = reactive<ShopUpdateReq>({
  shop_name: '',
  contact: '',
  phone: '',
  email: '',
  province: '',
  city: '',
  district: '',
  detail_address: '',
  address: '',
  remark: '',
})

const createRules: FormRules = {
  shop_code: [{ required: true, message: '请输入店铺编码', trigger: 'blur' }],
  shop_name: [{ required: true, message: '请输入店铺名称', trigger: 'blur' }],
  admin_username: [
    { required: true, message: '请输入管理员账户', trigger: 'blur' },
    { min: 3, message: '账户长度不能少于3位', trigger: 'blur' },
  ],
  admin_password: [
    { required: true, message: '请输入管理员密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' },
  ],
  email: [{ type: 'email', message: '请输入正确的邮箱地址', trigger: 'blur' }],
}

const editRules: FormRules = {
  shop_name: [{ required: true, message: '请输入店铺名称', trigger: 'blur' }],
  email: [{ type: 'email', message: '请输入正确的邮箱地址', trigger: 'blur' }],
}

function openCreateDialog() {
  isEdit.value = false
  dialogTitle.value = '新增店铺'
  editId.value = 0
  selectedRegion.value = []
  Object.assign(createForm, {
    shop_code: '',
    shop_name: '',
    contact: '',
    phone: '',
    email: '',
    province: '',
    city: '',
    district: '',
    detail_address: '',
    address: '',
    remark: '',
    admin_username: '',
    admin_password: '',
    admin_real_name: '',
  })
  dialogVisible.value = true
}

async function openEditDialog(row: ShopResp) {
  isEdit.value = true
  dialogTitle.value = '编辑店铺'
  editId.value = row.id
  const res = await getShop(row.id)
  const shop = res.data.data
  Object.assign(editForm, {
    shop_name: shop.shop_name,
    contact: shop.contact,
    phone: shop.phone,
    email: shop.email,
    province: shop.province,
    city: shop.city,
    district: shop.district,
    detail_address: shop.detail_address,
    address: shop.address,
    remark: shop.remark,
  })
  selectedRegion.value = findRegionCodes(shop.province, shop.city, shop.district)
  dialogVisible.value = true
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  submitLoading.value = true
  try {
    if (isEdit.value) {
      // Populate region names from cascader codes
      if (selectedRegion.value.length > 0) {
        editForm.province = codeToText[selectedRegion.value[0]] ?? ''
        editForm.city = selectedRegion.value[1] ? (codeToText[selectedRegion.value[1]] ?? '') : ''
        editForm.district = selectedRegion.value[2] ? (codeToText[selectedRegion.value[2]] ?? '') : ''
      }
      await updateShop(editId.value, editForm)
      ElMessage.success('更新成功')
    } else {
      // Populate region names from cascader codes
      if (selectedRegion.value.length > 0) {
        createForm.province = codeToText[selectedRegion.value[0]] ?? ''
        createForm.city = selectedRegion.value[1] ? (codeToText[selectedRegion.value[1]] ?? '') : ''
        createForm.district = selectedRegion.value[2] ? (codeToText[selectedRegion.value[2]] ?? '') : ''
      }
      await createShop(createForm)
      ElMessage.success('创建成功，店铺管理员账户已自动初始化')
    }
    dialogVisible.value = false
    fetchList()
  } finally {
    submitLoading.value = false
  }
}

async function handleDelete(row: ShopResp) {
  await ElMessageBox.confirm(
    `确定要删除店铺「${row.shop_name}」吗？此操作不可恢复。`,
    '删除确认',
    { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
  )
  await deleteShop(row.id)
  ElMessage.success('删除成功')
  fetchList()
}

async function handleToggleStatus(row: ShopResp) {
  const next = row.status === 1 ? 2 : 1
  const action = next === 1 ? '启用' : '停用'
  await ElMessageBox.confirm(
    `确定要${action}店铺「${row.shop_name}」吗？`,
    `${action}确认`,
    { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
  )
  await updateShopStatus(row.id, { status: next })
  ElMessage.success(`${action}成功`)
  fetchList()
}

const pwdDialogVisible = ref(false)
const pwdShopId = ref(0)
const pwdShopName = ref('')
const pwdNewPassword = ref('')
const pwdLoading = ref(false)
const pwdFormRef = ref<FormInstance>()
const pwdRules: FormRules = {
  new_password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' },
  ],
}

function openResetPwdDialog(row: ShopResp) {
  pwdShopId.value = row.id
  pwdShopName.value = row.shop_name
  pwdNewPassword.value = ''
  pwdDialogVisible.value = true
}

async function handleResetPwdSubmit() {
  const valid = await pwdFormRef.value?.validate().catch(() => false)
  if (!valid) return
  pwdLoading.value = true
  try {
    await resetShopAdminPassword(pwdShopId.value, pwdNewPassword.value)
    ElMessage.success(`店铺「${pwdShopName.value}」的管理员密码已重置`)
    pwdDialogVisible.value = false
  } finally {
    pwdLoading.value = false
  }
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
  <div class="shop-page">
    <el-card shadow="never" class="search-card">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="店铺编码">
          <el-input v-model="searchForm.shop_code" placeholder="请输入店铺编码" clearable />
        </el-form-item>
        <el-form-item label="店铺名称">
          <el-input v-model="searchForm.shop_name" placeholder="请输入店铺名称" clearable />
        </el-form-item>
        <el-form-item label="所在地区">
          <el-cascader
            v-model="searchRegion"
            :options="regionData"
            :props="{ expandTrigger: 'hover', value: 'value', label: 'label', children: 'children' }"
            placeholder="省/市/区"
            clearable
            style="width: 240px"
          />
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
        <el-button v-permission="'platform:shop:create'" type="primary" @click="openCreateDialog">
          新增店铺
        </el-button>
      </div>

      <el-table v-loading="loading" :data="tableData" border stripe style="width: 100%">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="shop_code" label="店铺编码" min-width="140" />
        <el-table-column prop="shop_name" label="店铺名称" min-width="160" />
        <el-table-column prop="contact" label="联系人" min-width="100" />
        <el-table-column prop="phone" label="联系电话" min-width="130" />
        <el-table-column label="地址" min-width="240">
          <template #default="{ row }">
            {{ [row.province, row.city, row.district].filter(Boolean).join(' ') }}{{ row.detail_address ? ' ' + row.detail_address : '' }}
          </template>
        </el-table-column>
        <el-table-column prop="admin_username" label="管理员账户" min-width="140" />
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
        <el-table-column label="操作" width="320" fixed="right">
          <template #default="{ row }">
            <el-button
              v-permission="'platform:shop:update'"
              type="primary" link size="small"
              @click="openEditDialog(row)"
            >
              编辑
            </el-button>
            <el-button
              v-permission="'platform:shop:status'"
              :type="row.status === 1 ? 'warning' : 'success'" link size="small"
              @click="handleToggleStatus(row)"
            >
              {{ row.status === 1 ? '停用' : '启用' }}
            </el-button>
            <el-button
              v-permission="'platform:shop:reset'"
              type="info" link size="small"
              @click="openResetPwdDialog(row)"
            >
              重置密码
            </el-button>
            <el-button
              v-permission="'platform:shop:delete'"
              type="danger" link size="small"
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

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="640px" :close-on-click-modal="false" destroy-on-close>
      <el-form
        v-if="!isEdit"
        ref="formRef" :model="createForm" :rules="createRules"
        label-width="120px"
      >
        <el-form-item label="店铺编码" prop="shop_code">
          <el-input v-model="createForm.shop_code" placeholder="请输入店铺编码（唯一）" />
        </el-form-item>
        <el-form-item label="店铺名称" prop="shop_name">
          <el-input v-model="createForm.shop_name" placeholder="请输入店铺名称" />
        </el-form-item>
        <el-form-item label="联系人">
          <el-input v-model="createForm.contact" placeholder="请输入联系人" />
        </el-form-item>
        <el-form-item label="联系电话">
          <el-input v-model="createForm.phone" placeholder="请输入联系电话" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="createForm.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="所在地区" required>
          <el-cascader
            v-model="selectedRegion"
            :options="regionData"
            :props="{ expandTrigger: 'hover', value: 'value', label: 'label', children: 'children' }"
            placeholder="请选择省/市/区"
            clearable
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="详细地址">
          <el-input v-model="createForm.detail_address" placeholder="请输入街道、门牌号等详细信息" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="createForm.remark" type="textarea" :rows="2" placeholder="可选" />
        </el-form-item>
        <el-divider content-position="left">店铺管理员账户（创建时自动初始化）</el-divider>
        <el-form-item label="管理员账户" prop="admin_username">
          <el-input v-model="createForm.admin_username" placeholder="请输入管理员登录账户" />
        </el-form-item>
        <el-form-item label="管理员密码" prop="admin_password">
          <el-input
            v-model="createForm.admin_password" type="password" show-password
            placeholder="请输入管理员初始密码（至少6位）"
          />
        </el-form-item>
        <el-form-item label="管理员姓名">
          <el-input
            v-model="createForm.admin_real_name" placeholder="留空则默认为「店长」"
          />
        </el-form-item>
      </el-form>

      <el-form
        v-else
        ref="formRef" :model="editForm" :rules="editRules" label-width="120px"
      >
        <el-form-item label="店铺名称" prop="shop_name">
          <el-input v-model="editForm.shop_name" placeholder="请输入店铺名称" />
        </el-form-item>
        <el-form-item label="联系人">
          <el-input v-model="editForm.contact" placeholder="请输入联系人" />
        </el-form-item>
        <el-form-item label="联系电话">
          <el-input v-model="editForm.phone" placeholder="请输入联系电话" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="editForm.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="所在地区" required>
          <el-cascader
            v-model="selectedRegion"
            :options="regionData"
            :props="{ expandTrigger: 'hover', value: 'value', label: 'label', children: 'children' }"
            placeholder="请选择省/市/区"
            clearable
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="详细地址">
          <el-input v-model="editForm.detail_address" placeholder="请输入街道、门牌号等详细信息" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="editForm.remark" type="textarea" :rows="2" placeholder="可选" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="pwdDialogVisible" title="重置管理员密码" width="440px" :close-on-click-modal="false" destroy-on-close>
      <el-form
        ref="pwdFormRef" :model="{ new_password: pwdNewPassword }" :rules="pwdRules"
        label-width="100px"
      >
        <p style="margin: 0 0 12px 0; color: #909399">
          将重置店铺「<strong>{{ pwdShopName }}</strong>」的管理员密码
        </p>
        <el-form-item label="新密码" prop="new_password">
          <el-input
            v-model="pwdNewPassword" type="password" show-password
            placeholder="请输入新密码（至少6位）"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="pwdDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="pwdLoading" @click="handleResetPwdSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.shop-page { padding: 16px; }
.search-card { margin-bottom: 16px; }
.search-card :deep(.el-card__body) { padding-bottom: 2px; }
.table-header { margin-bottom: 16px; }
.pagination-wrap { display: flex; justify-content: flex-end; margin-top: 16px; }
</style>
