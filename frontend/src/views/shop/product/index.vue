<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { ElTable } from 'element-plus'
import {
  getShopProducts,
  getPlatformProducts,
  selectProducts,
  updateShopProductPrice,
  updateShopProductStatus,
  cancelShopProduct,
} from '@/api/shop/product'
import type {
  ShopProductResp,
  ShopPlatformProductResp,
} from '@/types/system'

const loading = ref(false)
const tableData = ref<ShopProductResp[]>([])
const total = ref(0)

const searchForm = reactive({
  product_name: '',
  status: null as number | null,
})

const pagination = reactive({ page: 1, page_size: 10 })

const formatPrice = (val: number | null): string => {
  if (val == null) return '-'
  return val.toFixed(2)
}

const formatTime = (val: string | null): string => {
  if (!val) return '-'
  return val.replace('T', ' ').substring(0, 19)
}

const statusTagType = (status: number) => (status === 1 ? 'success' : 'info')
const statusText = (status: number) => (status === 1 ? '上架' : '下架')

async function fetchList() {
  loading.value = true
  try {
    const res = await getShopProducts({
      page: pagination.page,
      page_size: pagination.page_size,
      product_name: searchForm.product_name || undefined,
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
  searchForm.product_name = ''
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

const selectDialogVisible = ref(false)
const selectDialogLoading = ref(false)
const platformProducts = ref<ShopPlatformProductResp[]>([])
const selectedPlatformIds = ref<Set<number>>(new Set())
const selectTableRef = ref<InstanceType<typeof ElTable>>()
const dialogSearchKeyword = ref('')
const selectSubmitLoading = ref(false)

const filteredPlatformProducts = () => {
  if (!dialogSearchKeyword.value) return platformProducts.value
  const kw = dialogSearchKeyword.value.toLowerCase()
  return platformProducts.value.filter(
    (p) =>
      p.product_name.toLowerCase().includes(kw) ||
      p.product_code.toLowerCase().includes(kw),
  )
}

const isAlreadySelected = (row: ShopPlatformProductResp) => {
  return selectedPlatformIds.value.has(row.id)
}

function handleSelectDialogSelectionChange() {}

async function openSelectDialog() {
  selectDialogVisible.value = true
  selectDialogLoading.value = true
  dialogSearchKeyword.value = ''
  try {
    const [platformRes, shopRes] = await Promise.all([
      getPlatformProducts(),
      getShopProducts({ page: 1, page_size: 9999 }),
    ])
    platformProducts.value = platformRes.data.data
    const shopList: ShopProductResp[] = shopRes.data.data.list
    selectedPlatformIds.value = new Set(shopList.map((item) => item.platform_product_id))
  } finally {
    selectDialogLoading.value = false
  }
}

async function handleSelectSubmit() {
  const checkedRows = selectTableRef.value?.getSelectionRows() as ShopPlatformProductResp[] | undefined
  if (!checkedRows || checkedRows.length === 0) {
    ElMessage.warning('请选择要添加的商品')
    return
  }
  const newIds = checkedRows
    .filter((row) => !selectedPlatformIds.value.has(row.id))
    .map((row) => row.id)
  if (newIds.length === 0) {
    ElMessage.warning('所选商品均已添加，请选择新商品')
    return
  }
  selectSubmitLoading.value = true
  try {
    await selectProducts({ platform_product_ids: newIds })
    ElMessage.success('选品成功')
    selectDialogVisible.value = false
    fetchList()
  } finally {
    selectSubmitLoading.value = false
  }
}

const priceDialogVisible = ref(false)
const priceSubmitLoading = ref(false)
const priceForm = reactive({
  id: 0,
  shop_price: 0,
})

function openPriceDialog(row: ShopProductResp) {
  priceForm.id = row.id
  priceForm.shop_price = row.shop_price
  priceDialogVisible.value = true
}

async function handlePriceSubmit() {
  priceSubmitLoading.value = true
  try {
    await updateShopProductPrice(priceForm.id, { shop_price: priceForm.shop_price })
    ElMessage.success('改价成功')
    priceDialogVisible.value = false
    fetchList()
  } finally {
    priceSubmitLoading.value = false
  }
}

async function handleToggleStatus(row: ShopProductResp) {
  const next = row.status === 1 ? 2 : 1
  const action = next === 1 ? '上架' : '下架'
  await ElMessageBox.confirm(
    `确定要${action}商品「${row.product_name}」吗？`,
    `${action}确认`,
    { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' },
  )
  await updateShopProductStatus(row.id, { status: next })
  ElMessage.success(`${action}成功`)
  fetchList()
}

async function handleCancelSelect(row: ShopProductResp) {
  await ElMessageBox.confirm(
    `确定要取消选品「${row.product_name}」吗？`,
    '取消确认',
    { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' },
  )
  await cancelShopProduct(row.id)
  ElMessage.success('取消选品成功')
  fetchList()
}

onMounted(() => {
  fetchList()
})
</script>

<template>
  <div class="shop-product-page">
    <el-card shadow="never" class="search-card">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="商品名称">
          <el-input v-model="searchForm.product_name" placeholder="请输入商品名称" clearable />
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
        <el-button v-permission="'shop:product:select'" type="primary" @click="openSelectDialog">
          选品
        </el-button>
      </div>

      <el-table v-loading="loading" :data="tableData" border stripe style="width: 100%">
        <el-table-column prop="id" label="序号" width="70" />
        <el-table-column prop="product_code" label="商品编号" min-width="130" />
        <el-table-column prop="product_name" label="商品名称" min-width="140" />
        <el-table-column label="平台价格" width="100" align="right">
          <template #default="{ row }">
            {{ formatPrice(row.platform_price) }}
          </template>
        </el-table-column>
        <el-table-column label="店铺售价" width="100" align="right">
          <template #default="{ row }">
            {{ formatPrice(row.shop_price) }}
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
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button
              v-permission="'shop:product:price'"
              type="primary" link size="small"
              @click="openPriceDialog(row)"
            >
              改价
            </el-button>
            <el-button
              v-permission="'shop:product:status'"
              :type="row.status === 1 ? 'warning' : 'success'" link size="small"
              @click="handleToggleStatus(row)"
            >
              {{ row.status === 1 ? '下架' : '上架' }}
            </el-button>
            <el-button
              v-permission="'shop:product:delete'"
              type="danger" link size="small"
              @click="handleCancelSelect(row)"
            >
              取消选品
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
      v-model="selectDialogVisible"
      title="选品"
      width="800px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <div v-loading="selectDialogLoading">
        <div class="dialog-search">
          <el-input
            v-model="dialogSearchKeyword"
            placeholder="搜索商品名称/编号"
            clearable
            style="width: 260px"
          />
        </div>
        <el-table
          ref="selectTableRef"
          :data="filteredPlatformProducts()"
          border
          style="width: 100%"
          max-height="400"
          @selection-change="handleSelectDialogSelectionChange"
        >
          <el-table-column type="selection" width="50" :selectable="(row: ShopPlatformProductResp) => !isAlreadySelected(row)" />
          <el-table-column prop="product_code" label="商品编号" min-width="130" />
          <el-table-column prop="product_name" label="商品名称" min-width="140" />
          <el-table-column label="分类" min-width="100">
            <template #default="{ row }">
              {{ row.category_name || '-' }}
            </template>
          </el-table-column>
          <el-table-column label="平台价格" width="100" align="right">
            <template #default="{ row }">
              {{ formatPrice(row.price) }}
            </template>
          </el-table-column>
          <el-table-column prop="description" label="描述" min-width="150" show-overflow-tooltip />
        </el-table>
      </div>
      <template #footer>
        <el-button @click="selectDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="selectSubmitLoading" @click="handleSelectSubmit">确认选品</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="priceDialogVisible"
      title="改价"
      width="400px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-form label-width="80px">
        <el-form-item label="店铺售价">
          <el-input-number
            v-model="priceForm.shop_price"
            :precision="2"
            :min="0"
            :step="10"
            style="width: 100%"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="priceDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="priceSubmitLoading" @click="handlePriceSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.shop-product-page { padding: 16px; }
.search-card { margin-bottom: 16px; }
.search-card :deep(.el-card__body) { padding-bottom: 2px; }
.table-header { margin-bottom: 16px; }
.pagination-wrap { display: flex; justify-content: flex-end; margin-top: 16px; }
.dialog-search { margin-bottom: 12px; }
</style>
