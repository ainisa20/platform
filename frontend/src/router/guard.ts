import type { Router } from 'vue-router'
import { getToken } from '@/utils/auth'
import { useUserStore } from '@/store/user'
import { usePermissionStore } from '@/store/permission'
import platformRoutes from './modules/platform'
import shopRoutes from './modules/shop'

const WHITE_LIST = ['/platform/login', '/shop/login']

function detectSystem(path: string): 'platform' | 'shop' {
  if (path.startsWith('/shop')) return 'shop'
  return 'platform'
}

export function setupRouterGuard(router: Router) {
  router.beforeEach(async (to, _from, next) => {
    document.title = (to.meta.title as string) || 'SaaS管理平台'

    if (WHITE_LIST.includes(to.path)) {
      next()
      return
    }

    const token = getToken()
    if (!token) {
      const loginPath = detectSystem(to.path) === 'shop' ? '/shop/login' : '/platform/login'
      next({ path: loginPath, query: { redirect: to.fullPath } })
      return
    }

    const userStore = useUserStore()

    if (!userStore.userInfo) {
      try {
        await userStore.getUserInfo()
        const permissionStore = usePermissionStore()
        const codes = (userStore.userInfo as any)?.permissions as string[] | undefined
        permissionStore.setPermissionCodes(codes || [])

        const systemRoutes = detectSystem(to.path) === 'platform' ? platformRoutes : shopRoutes
        permissionStore.generateRoutes(systemRoutes)

        next({ ...to, replace: true })
        return
      } catch {
        userStore.resetState()
        const loginPath = detectSystem(to.path) === 'shop' ? '/shop/login' : '/platform/login'
        next({ path: loginPath, query: { redirect: to.fullPath } })
        return
      }
    }

    next()
  })
}
