<script setup lang="ts">
import { ref, reactive, computed, onMounted, nextTick, watch } from 'vue'
import { ElMessage } from 'element-plus'
import * as echarts from 'echarts'
import {
  getPlatformFinanceSummary,
  getPlatformFinanceTrend,
  getPlatformFinanceProfitLoss,
  getPlatformFinancePerShop,
} from '@/api/platform/report'
import { getShopList } from '@/api/platform/shop'
import type {
  FinanceSummaryResp,
  FinanceTrendItem,
  ProfitLossCategory,
  FinanceReportShopSummary,
  ShopResp,
} from '@/types/system'

const loading = ref(false)
const shops = ref<ShopResp[]>([])
const selectedShopId = ref<number | null>(null)
const dateRange = ref<[string, string]>(['', ''])

const summary = ref<FinanceSummaryResp>({ total_income: 0, total_expense: 0, net_profit: 0 })
const trendData = ref<FinanceTrendItem[]>([])
const profitLossData = ref<ProfitLossCategory[]>([])
const perShopData = ref<FinanceReportShopSummary[]>([])
const trendChart = ref<HTMLElement | null>(null)
const shopChart = ref<HTMLElement | null>(null)
let trendChartInstance: echarts.ECharts | null = null
let shopChartInstance: echarts.ECharts | null = null

const queryParams = computed(() => ({
  shop_id: selectedShopId.value ?? undefined,
  start_date: dateRange.value[0] || undefined,
  end_date: dateRange.value[1] || undefined,
}))

onMounted(async () => {
  await loadShops()
  await fetchAll()
})

watch(selectedShopId, () => fetchAll())
watch(dateRange, () => fetchAll(), { deep: true })

async function loadShops() {
  try {
    const res = await getShopList({ page: 1, page_size: 200 })
    shops.value = res.data.data.list
  } catch {
    ElMessage.error('加载店铺列表失败')
  }
}

async function fetchAll() {
  loading.value = true
  try {
    await Promise.all([
      fetchSummary(),
      fetchTrend(),
      fetchProfitLoss(),
      fetchPerShop(),
    ])
  } finally {
    loading.value = false
  }
}

async function fetchSummary() {
  const res = await getPlatformFinanceSummary(queryParams.value)
  summary.value = res.data.data
}

async function fetchTrend() {
  const res = await getPlatformFinanceTrend({ ...queryParams.value, months: 12 })
  trendData.value = res.data.data
  renderTrendChart()
}

async function fetchProfitLoss() {
  const res = await getPlatformFinanceProfitLoss(queryParams.value)
  profitLossData.value = res.data.data.categories
}

async function fetchPerShop() {
  if (selectedShopId.value) {
    perShopData.value = []
    return
  }
  const res = await getPlatformFinancePerShop({
    start_date: dateRange.value[0] || undefined,
    end_date: dateRange.value[1] || undefined,
  })
  perShopData.value = res.data.data
  await nextTick()
  renderShopChart()
}

function renderTrendChart() {
  if (!trendChart.value) return
  trendChartInstance?.dispose()
  trendChartInstance = echarts.init(trendChart.value)
  const months = trendData.value.map((d) => d.month)
  trendChartInstance.setOption({
    tooltip: { trigger: 'axis' },
    legend: { data: ['收入', '支出'] },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { type: 'category', data: months, boundaryGap: false },
    yAxis: { type: 'value' },
    series: [
      {
        name: '收入',
        type: 'line',
        smooth: true,
        data: trendData.value.map((d) => d.income),
        itemStyle: { color: '#409EFF' },
        areaStyle: { color: 'rgba(64,158,255,0.1)' },
      },
      {
        name: '支出',
        type: 'line',
        smooth: true,
        data: trendData.value.map((d) => d.expense),
        itemStyle: { color: '#F56C6C' },
        areaStyle: { color: 'rgba(245,108,108,0.1)' },
      },
    ],
  })
}

function renderShopChart() {
  if (!shopChart.value) return
  shopChartInstance?.dispose()
  shopChartInstance = echarts.init(shopChart.value)
  const names = perShopData.value.map((d) => d.shop_name)
  shopChartInstance.setOption({
    tooltip: { trigger: 'axis' },
    legend: { data: ['收入', '支出', '净利润'] },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { type: 'category', data: names },
    yAxis: { type: 'value' },
    series: [
      {
        name: '收入',
        type: 'bar',
        data: perShopData.value.map((d) => d.income),
        itemStyle: { color: '#409EFF' },
      },
      {
        name: '支出',
        type: 'bar',
        data: perShopData.value.map((d) => d.expense),
        itemStyle: { color: '#F56C6C' },
      },
      {
        name: '净利润',
        type: 'bar',
        data: perShopData.value.map((d) => d.net_profit),
        itemStyle: { color: '#67C23A' },
      },
    ],
  })
}

function formatMoney(val: number): string {
  return `¥${val.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
}

const incomeCategories = computed(() =>
  profitLossData.value.filter((c) => c.type === 'income'),
)
const expenseCategories = computed(() =>
  profitLossData.value.filter((c) => c.type === 'expense'),
)
</script>

<template>
  <div class="finance-report">
    <div class="filter-bar">
      <el-select
        v-model="selectedShopId"
        placeholder="全部店铺"
        clearable
        style="width: 200px"
      >
        <el-option label="全部店铺" :value="null" />
        <el-option
          v-for="shop in shops"
          :key="shop.id"
          :label="shop.shop_name"
          :value="shop.id"
        />
      </el-select>
      <el-date-picker
        v-model="dateRange"
        type="daterange"
        range-separator="至"
        start-placeholder="开始日期"
        end-placeholder="结束日期"
        value-format="YYYY-MM-DD"
        style="margin-left: 12px"
      />
    </div>

    <el-row :gutter="16" class="summary-cards">
      <el-col :span="8">
        <el-card shadow="hover">
          <div class="card-value income">{{ formatMoney(summary.total_income) }}</div>
          <div class="card-label">总收入</div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover">
          <div class="card-value expense">{{ formatMoney(summary.total_expense) }}</div>
          <div class="card-label">总支出</div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover">
          <div class="card-value" :class="summary.net_profit >= 0 ? 'profit' : 'loss'">
            {{ formatMoney(summary.net_profit) }}
          </div>
          <div class="card-label">净利润</div>
        </el-card>
      </el-col>
    </el-row>

    <el-card shadow="hover" class="chart-card">
      <template #header>
        <span>月度收支趋势</span>
      </template>
      <div ref="trendChart" style="height: 320px" v-loading="loading" />
    </el-card>

    <el-row :gutter="16">
      <el-col :span="12">
        <el-card shadow="hover" class="table-card">
          <template #header>
            <span>收入分类</span>
          </template>
          <el-table
            :data="incomeCategories"
            row-key="name"
            :tree-props="{ children: 'children' }"
            stripe
            size="small"
            v-loading="loading"
            default-expand-all
          >
            <el-table-column prop="name" label="分类" />
            <el-table-column prop="subtotal" label="金额" width="160">
              <template #default="{ row }">
                {{ formatMoney(row.subtotal) }}
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="hover" class="table-card">
          <template #header>
            <span>支出分类</span>
          </template>
          <el-table
            :data="expenseCategories"
            row-key="name"
            :tree-props="{ children: 'children' }"
            stripe
            size="small"
            v-loading="loading"
            default-expand-all
          >
            <el-table-column prop="name" label="分类" />
            <el-table-column prop="subtotal" label="金额" width="160">
              <template #default="{ row }">
                {{ formatMoney(row.subtotal) }}
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>

    <el-card v-if="!selectedShopId && perShopData.length > 0" shadow="hover" class="chart-card">
      <template #header>
        <span>各店铺收支对比</span>
      </template>
      <div ref="shopChart" style="height: 320px" v-loading="loading" />
    </el-card>
  </div>
</template>

<style scoped>
.finance-report {
  padding: 16px;
}
.filter-bar {
  margin-bottom: 16px;
  display: flex;
  align-items: center;
}
.summary-cards {
  margin-bottom: 16px;
}
.card-value {
  font-size: 28px;
  font-weight: 700;
  line-height: 1.4;
}
.card-label {
  font-size: 14px;
  color: #909399;
  margin-top: 4px;
}
.income {
  color: #409eff;
}
.expense {
  color: #f56c6c;
}
.profit {
  color: #67c23a;
}
.loss {
  color: #f56c6c;
}
.chart-card {
  margin-bottom: 16px;
}
.table-card {
  margin-bottom: 16px;
}
</style>
