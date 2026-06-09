<script setup lang="ts">
import { ref, nextTick, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { ElTree } from 'element-plus'
import {
  getShopFinCategories,
  getAvailableFinCategories,
  syncFinCategories,
  cancelSyncFinCategory,
} from '@/api/shop/finance'
import type {
  ShopFinCategoryResp,
  ShopFinCategoryAvailableResp,
} from '@/types/system'

const loading = ref(false)
const tableData = ref<ShopFinCategoryResp[]>([])
const activeType = ref<number | null>(null)

const syncDialogVisible = ref(false)
const syncLoading = ref(false)
const availableData = ref<ShopFinCategoryAvailableResp[]>([])
const syncedPlatformIds = ref<Set<number>>(new Set())
const syncTreeRef = ref<InstanceType<typeof ElTree>>()
const submitSyncLoading = ref(false)

async function fetchList() {
  loading.value = true
  try {
    const res = await getShopFinCategories({ category_type: activeType.value })
    tableData.value = res.data.data
  } finally {
    loading.value = false
  }
}

function handleTypeChange() {
  fetchList()
}

function collectPlatformIds(nodes: ShopFinCategoryResp[]): number[] {
  const ids: number[] = []
  for (const n of nodes) {
    ids.push(n.platform_category_id)
    if (n.children) ids.push(...collectPlatformIds(n.children))
  }
  return ids
}

async function openSyncDialog() {
  syncDialogVisible.value = true
  syncLoading.value = true
  try {
    const [availableRes, syncedRes] = await Promise.all([
      getAvailableFinCategories(),
      getShopFinCategories(),
    ])
    availableData.value = availableRes.data.data
    syncedPlatformIds.value = new Set(collectPlatformIds(syncedRes.data.data))
    await nextTick()
    if (syncTreeRef.value) {
      syncTreeRef.value.setCheckedKeys(Array.from(syncedPlatformIds.value))
    }
  } finally {
    syncLoading.value = false
  }
}

async function handleSubmitSync() {
  const checked = syncTreeRef.value?.getCheckedNodes(false, true) as ShopFinCategoryAvailableResp[]
  const newIds = checked
    .filter(n => !syncedPlatformIds.value.has(n.id))
    .map(n => n.id)

  if (newIds.length === 0) {
    ElMessage.warning('没有新的分类需要同步')
    return
  }

  submitSyncLoading.value = true
  try {
    await syncFinCategories({ platform_category_ids: newIds })
    ElMessage.success('同步成功')
    syncDialogVisible.value = false
    fetchList()
  } finally {
    submitSyncLoading.value = false
  }
}

async function handleCancelSync(row: ShopFinCategoryResp) {
  await ElMessageBox.confirm(
    `确定要取消同步分类「${row.category_name}」吗？`,
    '取消同步确认',
    { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
  )
  await cancelSyncFinCategory(row.id)
  ElMessage.success('已取消同步')
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
    <el-card shadow="never" class="table-card">
      <div class="table-header">
        <el-radio-group v-model="activeType" @change="handleTypeChange" style="margin-right: 16px">
          <el-radio-button :value="null">全部</el-radio-button>
          <el-radio-button :value="1">收入</el-radio-button>
          <el-radio-button :value="2">支出</el-radio-button>
        </el-radio-group>
        <el-button v-permission="'shop:finance:category:sync'" type="primary" @click="openSyncDialog">
          同步分类
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
        <el-table-column prop="level" label="层级" width="80" align="center" />
        <el-table-column label="同步时间" min-width="170">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button
              v-permission="'shop:finance:category:delete'"
              type="danger"
              link
              size="small"
              @click="handleCancelSync(row)"
            >
              取消同步
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog
      v-model="syncDialogVisible"
      title="同步分类"
      width="600px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <div v-loading="syncLoading">
        <el-tree
          ref="syncTreeRef"
          :data="availableData"
          show-checkbox
          check-strictly
          node-key="id"
          :props="{ label: 'category_name', children: 'children' }"
          default-expand-all
        />
      </div>
      <template #footer>
        <el-button @click="syncDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitSyncLoading" @click="handleSubmitSync">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.category-page { padding: 16px; }
.table-card { margin-bottom: 16px; }
.table-header { display: flex; align-items: center; margin-bottom: 16px; }
</style>
