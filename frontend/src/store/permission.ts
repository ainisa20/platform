import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getSystemType } from '@/utils/auth'
import * as platformAuth from '@/api/platform/auth'
import * as shopAuth from '@/api/shop/auth'

export const usePermissionStore = defineStore('permission', () => {
  const permissionCodes = ref<string[]>([])
  const menuRoutes = ref<any[]>([])

  async function fetchPermissions() {
    const authApi = getSystemType() === 'platform' ? platformAuth : shopAuth
    const res = await authApi.getPermissions()
    permissionCodes.value = res.data.data || []
    return permissionCodes.value
  }

  function setPermissionCodes(codes: string[]) {
    permissionCodes.value = codes
  }

  function generateRoutes(routes: any[]) {
    const filtered = filterRoutes(routes)
    menuRoutes.value = filtered
    return filtered
  }

  function filterRoutes(routes: any[]): any[] {
    return routes
      .map((route) => {
        // Deep clone to avoid mutating original route definitions
        const clone = { ...route }
        if (clone.children) {
          clone.children = filterRoutes(clone.children)
        }
        return clone
      })
      .filter((route) => {
        if (route.meta?.permissionCode) {
          return permissionCodes.value.includes(route.meta.permissionCode)
        }
        // Parent routes without permissionCode: show only if they have visible children
        if (route.children && route.children.length > 0) {
          return true
        }
        return true
      })
      .filter((route) => {
        // Hide parent routes whose children were ALL filtered out
        if (!route.meta?.permissionCode && route.children && route.children.length === 0) {
          return false
        }
        return true
      })
  }

  function resetState() {
    permissionCodes.value = []
    menuRoutes.value = []
  }

  return {
    permissionCodes,
    menuRoutes,
    fetchPermissions,
    setPermissionCodes,
    generateRoutes,
    resetState,
  }
})
