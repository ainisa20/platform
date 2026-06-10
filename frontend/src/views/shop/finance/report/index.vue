<template>
  <div class="page-container">
    <el-card>
      <template #header>
        <span class="page-title">财务报表</span>
      </template>

      <el-tabs v-model="activeTab">
        <el-tab-pane label="收支汇总" name="summary">
          <div class="tab-toolbar">
            <el-date-picker
              v-model="summaryRange"
              type="daterange"
              range-separator="至"
              start-placeholder="开始日期"
              end-placeholder="结束日期"
              value-format="YYYY-MM-DD"
              style="width: 280px"
              @change="fetchSummary"
            />
          </div>

          <div class="stat-cards" v-loading="summaryLoading">
            <div class="stat-card stat-income">
              <div class="stat-label">收入合计</div>
              <div class="stat-value">¥{{ formatMoney(summary.total_income) }}</div>
            </div>
            <div class="stat-card stat-expense">
              <div class="stat-label">支出合计</div>
              <div class="stat-value">¥{{ formatMoney(summary.total_expense) }}</div>
            </div>
            <div class="stat-card stat-profit">
              <div class="stat-label">净收入</div>
              <div class="stat-value">¥{{ formatMoney(summary.net_profit) }}</div>
            </div>
            <div class="stat-card stat-rate">
              <div class="stat-label">利润率</div>
              <div class="stat-value">{{ profitRate }}</div>
            </div>
          </div>

          <el-row :gutter="16" class="chart-row">
            <el-col :span="14">
              <el-card shadow="hover">
                <template #header>
                  <span class="card-title">收支趋势对比</span>
                </template>
                <div ref="trendChartRef" class="trend-chart" />
              </el-card>
            </el-col>
            <el-col :span="10">
              <el-card shadow="hover">
                <template #header>
                  <span class="card-title">支出分类占比</span>
                </template>
                <div ref="pieChartRef" class="trend-chart" />
              </el-card>
            </el-col>
          </el-row>
        </el-tab-pane>

        <el-tab-pane label="利润表" name="profit">
          <div class="tab-toolbar">
            <el-date-picker
              v-model="profitRange"
              type="daterange"
              range-separator="至"
              start-placeholder="开始日期"
              end-placeholder="结束日期"
              value-format="YYYY-MM-DD"
              style="width: 280px"
              @change="fetchProfitLoss"
            />
          </div>

          <el-table :data="profitTableData" stripe style="width: 100%">
            <el-table-column prop="item" label="项目" min-width="200" />
            <el-table-column label="金额" min-width="160" align="right">
              <template #default="{ row }">
                <span :class="row.amountClass">
                  ¥{{ formatMoney(row.amount) }}
                </span>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick, watch, shallowRef } from 'vue'
import { ElMessage } from 'element-plus'
import { getFinanceSummary, getFinanceTrend, getFinanceProfitLoss } from '@/api/shop/report'
import type { FinanceTrendItem, ProfitLossCategory } from '@/types/system'
import * as echarts from 'echarts'

const activeTab = ref('summary')

const summaryRange = ref<[string, string] | null>(null)
const summaryLoading = ref(false)
const summary = ref({
  total_income: 0,
  total_expense: 0,
  net_profit: 0,
})

const profitRate = computed(() => {
  const income = Number(summary.value.total_income)
  if (!income) return '0%'
  return `${((Number(summary.value.net_profit) / income) * 100).toFixed(1)}%`
})

const profitRange = ref<[string, string] | null>(null)
const profitLossRaw = ref<ProfitLossCategory[]>([])

const trendData = ref<FinanceTrendItem[]>([])

interface ProfitTableRow {
  item: string
  amount: number
  amountClass: string
}

const profitTableData = computed<ProfitTableRow[]>(() => {
  const incomeRows = profitLossRaw.value.filter((r) => r.type === 'income')
  const expenseRows = profitLossRaw.value.filter((r) => r.type === 'expense')
  const incomeTotal = incomeRows.reduce((s, r) => s + Number(r.subtotal || 0), 0)
  const expenseTotal = expenseRows.reduce((s, r) => s + Number(r.subtotal || 0), 0)
  const net = incomeTotal - expenseTotal

  return [
    { item: '一、收入', amount: 0, amountClass: 'header' },
    ...incomeRows.map((r) => ({
      item: '  ' + r.name,
      amount: Number(r.subtotal || 0),
      amountClass: 'text-green',
    })),
    { item: '收入小计', amount: incomeTotal, amountClass: 'subtotal' },
    { item: '二、支出', amount: 0, amountClass: 'header' },
    ...expenseRows.map((r) => ({
      item: '  ' + r.name,
      amount: -Number(r.subtotal || 0),
      amountClass: 'text-red',
    })),
    { item: '支出小计', amount: -expenseTotal, amountClass: 'subtotal' },
    { item: '三、净收入', amount: net, amountClass: 'subtotal' },
  ]
})

const trendChartRef = ref<HTMLElement | null>(null)
const pieChartRef = ref<HTMLElement | null>(null)
const trendChart = shallowRef<echarts.ECharts>()
const pieChart = shallowRef<echarts.ECharts>()

function formatMoney(val: number): string {
  return (val ?? 0).toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function monthLabel(ym: string): string {
  const parts = ym.split('-')
  return parts.length === 2 ? `${parseInt(parts[1], 10)}月` : ym
}

async function fetchSummary() {
  summaryLoading.value = true
  try {
    const [summaryRes, trendRes, plRes] = await Promise.all([
      getFinanceSummary(summaryRange.value?.[0], summaryRange.value?.[1]),
      getFinanceTrend(6),
      getFinanceProfitLoss(summaryRange.value?.[0], summaryRange.value?.[1]),
    ])
    const sd = summaryRes.data.data
    summary.value = {
      total_income: sd?.total_income ?? 0,
      total_expense: sd?.total_expense ?? 0,
      net_profit: sd?.net_profit ?? 0,
    }
    const td = trendRes.data.data
    trendData.value = Array.isArray(td) ? td : []
    const pld = plRes.data.data
    profitLossRaw.value = Array.isArray(pld?.categories) ? pld.categories : []
    nextTick(() => {
      renderTrendChart()
      renderPieChart()
    })
  } catch {
    ElMessage.error('获取汇总数据失败')
  } finally {
    summaryLoading.value = false
  }
}

async function fetchProfitLoss() {
  try {
    const res = await getFinanceProfitLoss(profitRange.value?.[0], profitRange.value?.[1])
    const d = res.data.data
    profitLossRaw.value = Array.isArray(d?.categories) ? d.categories : []
  } catch {
    ElMessage.error('获取利润表数据失败')
  }
}

function renderTrendChart() {
  if (!trendChartRef.value) return
  if (!trendChart.value) trendChart.value = echarts.init(trendChartRef.value)
  const months = trendData.value.map((d) => monthLabel(d.month))
  const incomeSeries = trendData.value.map((d) => Number(d.income || 0))
  const expenseSeries = trendData.value.map((d) => Number(d.expense || 0))
  trendChart.value.setOption(
    {
      tooltip: {
        trigger: 'axis',
        formatter: (params: { seriesName: string; value: number }[]) =>
          params
            .map((p) => `${p.seriesName}: ¥${p.value.toLocaleString('zh-CN', { minimumFractionDigits: 2 })}`)
            .join('<br/>'),
      },
      legend: { data: ['收入', '支出'], top: 10 },
      grid: { left: '3%', right: '4%', bottom: '3%', top: 50, containLabel: true },
      xAxis: { type: 'category', data: months, boundaryGap: false },
      yAxis: {
        type: 'value',
        axisLabel: {
          formatter: (val: number) => `¥${(val / 10000).toFixed(1)}万`,
        },
      },
      series: [
        {
          name: '收入',
          type: 'line',
          smooth: true,
          data: incomeSeries,
          lineStyle: { color: '#67c23a', width: 3 },
          itemStyle: { color: '#67c23a' },
          areaStyle: { color: 'rgba(103, 194, 58, 0.15)' },
        },
        {
          name: '支出',
          type: 'line',
          smooth: true,
          data: expenseSeries,
          lineStyle: { color: '#f56c6c', width: 3 },
          itemStyle: { color: '#f56c6c' },
          areaStyle: { color: 'rgba(245, 108, 108, 0.15)' },
        },
      ],
    },
    true,
  )
}

function renderPieChart() {
  if (!pieChartRef.value) return
  if (!pieChart.value) pieChart.value = echarts.init(pieChartRef.value)
  const expenseRows = profitLossRaw.value.filter((r) => r.type === 'expense' && Number(r.subtotal) > 0)
  const data = expenseRows.map((r) => ({ name: r.name, value: Number(r.subtotal) }))
  pieChart.value.setOption(
    {
      tooltip: { trigger: 'item', formatter: '{b}: ¥{c} ({d}%)' },
      legend: { bottom: 0 },
      series: [
        {
          type: 'pie',
          radius: ['40%', '68%'],
          center: ['50%', '45%'],
          label: { show: false },
          emphasis: { label: { show: true, fontWeight: 'bold' } },
          data,
        },
      ],
    },
    true,
  )
}

function handleResize() {
  trendChart.value?.resize()
  pieChart.value?.resize()
}

watch(activeTab, (val) => {
  if (val === 'summary') {
    nextTick(() => {
      if (trendChart.value) trendChart.value.resize()
      else renderTrendChart()
      if (pieChart.value) pieChart.value.resize()
      else renderPieChart()
    })
  }
})

onMounted(() => {
  fetchSummary()
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  trendChart.value?.dispose()
  pieChart.value?.dispose()
})
</script>

<style scoped>
.page-container {
  padding: 16px;
}

.page-title {
  font-size: 16px;
  font-weight: 600;
}

.tab-toolbar {
  margin-bottom: 20px;
}

.stat-cards {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
}

.stat-card {
  padding: 24px;
  border-radius: 8px;
  text-align: center;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  margin-bottom: 8px;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
}

.stat-income {
  background: #f0f9eb;
  border: 1px solid #e1f3d8;
}

.stat-income .stat-value {
  color: #67c23a;
}

.stat-expense {
  background: #fef0f0;
  border: 1px solid #fde2e2;
}

.stat-expense .stat-value {
  color: #f56c6c;
}

.stat-profit {
  background: #ecf5ff;
  border: 1px solid #d9ecff;
}

.stat-profit .stat-value {
  color: #409eff;
}

.stat-rate {
  background: #fdf6ec;
  border: 1px solid #faecd8;
}

.stat-rate .stat-value {
  color: #e6a23c;
}

.card-title {
  font-size: 15px;
  font-weight: 600;
  color: #303133;
}

.chart-row {
  margin-top: 16px;
}

.text-green {
  color: #67c23a;
  font-weight: 600;
}

.text-red {
  color: #f56c6c;
  font-weight: 600;
}

.header {
  color: #303133;
  font-weight: 700;
  background: #f5f7fa;
}

.subtotal {
  color: #303133;
  font-weight: 700;
  background: #fafafa;
}

.trend-chart {
  width: 100%;
  height: 350px;
}
</style>
