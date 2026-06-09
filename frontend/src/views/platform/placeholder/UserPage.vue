<script setup lang="ts">
import { useRoute } from 'vue-router'
const route = useRoute()
</script>

<template>
  <div class="placeholder">
    <el-card>
      <template #header>
        <span style="font-weight: 600; font-size: 16px">{{ route.meta?.title || '页面' }}</span>
      </template>
      <p style="color: #666">该页面计划在阶段二实现。当前阶段用于验证登录、权限和路由框架。</p>
      <p style="color: #999; font-size: 13px">权限码：<code>{{ route.meta?.permissionCode }}</code></p>
    </el-card>
  </div>
</template>

<style scoped>
.placeholder {
  padding: 20px;
}
</style>
