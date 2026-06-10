<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/store/user'
import { usePermissionStore } from '@/store/permission'
import { ElMessage } from 'element-plus'
import {
  Fold,
  Expand,
  User,
  Setting,
  Shop,
  Goods,
  Money,
  Key,
  OfficeBuilding,
  SwitchButton,
  List,
  Search,
  UserFilled,
  Document,
} from '@element-plus/icons-vue'

const iconMap: Record<string, any> = {
  User, Setting, Shop, Goods, Money, Key, OfficeBuilding, List, Search, UserFilled, Document,
}

const isCollapse = ref(false)
const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const permissionStore = usePermissionStore()

const systemTitle = computed(() => {
  return route.path.startsWith('/shop') ? '店铺管理系统' : '平台管理系统'
})

const activeMenu = computed(() => route.path)

const menuItems = computed(() => {
  const codes = permissionStore.permissionCodes
  const prefix = route.path.startsWith('/shop') ? '/shop/' : '/platform/'

  const topRoutes = router.getRoutes().filter((r) => {
    if (!r.path.startsWith(prefix) || !r.meta?.title || r.path.includes('login')) return false
    const parts = r.path.split('/').filter(Boolean)
    return parts.length === 2
  })

  return topRoutes
    .map((r) => {
      const rawChildren = r.children || []
      const hasSubRoutes = rawChildren.length > 0
      const visibleChildren = rawChildren
        .filter((c: any) => !c.meta?.permissionCode || codes.includes(c.meta.permissionCode))
        .map((c: any) => ({
          path: r.path + '/' + (c.path || ''),
          title: c.meta?.title as string,
          icon: c.meta?.icon as string,
        }))

      const permitted = !r.meta?.permissionCode || codes.includes(r.meta.permissionCode as string)

      if (hasSubRoutes && visibleChildren.length === 0) return null
      if (!permitted) return null

      return {
        path: r.redirect || r.path,
        title: r.meta?.title as string,
        icon: r.meta?.icon as string,
        children: visibleChildren,
      }
    })
    .filter(Boolean) as Array<{ path: string; title: string; icon: string; children: Array<{ path: string; title: string; icon: string }> }>
})

async function handleLogout() {
  await userStore.logout()
  ElMessage.success('已退出登录')
  const loginPath = route.path.startsWith('/shop') ? '/shop/login' : '/platform/login'
  router.push(loginPath)
}

function toggleCollapse() {
  isCollapse.value = !isCollapse.value
}
</script>

<template>
  <el-container class="layout-container">
    <el-aside :width="isCollapse ? '64px' : '220px'" class="layout-aside">
      <div class="logo-area">
        <span v-if="!isCollapse" class="logo-text">{{ systemTitle }}</span>
        <span v-else class="logo-text-mini">S</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        :collapse="isCollapse"
        :collapse-transition="false"
        router
        class="layout-menu"
        background-color="#001529"
        text-color="#ffffffa6"
        active-text-color="#ffffff"
      >
        <template v-for="item in menuItems" :key="item.path">
          <el-sub-menu v-if="item.children.length > 0" :index="item.path">
            <template #title>
              <el-icon><component :is="iconMap[item.icon] || Setting" /></el-icon>
              <span>{{ item.title }}</span>
            </template>
            <el-menu-item v-for="child in item.children" :key="child.path" :index="child.path">
              <el-icon><component :is="iconMap[child.icon] || Setting" /></el-icon>
              <span>{{ child.title }}</span>
            </el-menu-item>
          </el-sub-menu>
          <el-menu-item v-else :index="item.path">
            <el-icon><component :is="iconMap[item.icon] || Setting" /></el-icon>
            <span>{{ item.title }}</span>
          </el-menu-item>
        </template>
      </el-menu>
    </el-aside>

    <el-container>
      <el-header class="layout-header">
        <div class="header-left">
          <el-icon class="collapse-btn" @click="toggleCollapse">
            <Fold v-if="!isCollapse" />
            <Expand v-else />
          </el-icon>
        </div>
        <div class="header-right">
          <span class="user-name">{{ userStore.userInfo?.real_name || userStore.userInfo?.username || '' }}</span>
          <el-button :icon="SwitchButton" text @click="handleLogout">退出</el-button>
        </div>
      </el-header>
      <el-main class="layout-main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<style scoped>
.layout-container {
  height: 100vh;
}

.layout-aside {
  background-color: #001529;
  transition: width 0.28s;
  overflow: hidden;
}

.logo-area {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid #ffffff1a;
}

.logo-text {
  color: #fff;
  font-size: 16px;
  font-weight: 600;
  white-space: nowrap;
}

.logo-text-mini {
  color: #fff;
  font-size: 20px;
  font-weight: 700;
}

.layout-menu {
  border-right: none;
}

.layout-menu:not(.el-menu--collapse) {
  width: 220px;
}

.layout-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid #ebeef5;
  background: #fff;
  padding: 0 20px;
  height: 60px;
}

.header-left {
  display: flex;
  align-items: center;
}

.collapse-btn {
  font-size: 20px;
  cursor: pointer;
  color: #333;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-name {
  font-size: 14px;
  color: #333;
}

.layout-main {
  background-color: #f5f5f5;
  min-height: 0;
}
</style>
