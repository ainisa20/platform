import type { App, Directive, DirectiveBinding } from 'vue'
import { usePermissionStore } from '@/store/permission'

export const permissionDirective: Directive = {
  mounted(el: HTMLElement, binding: DirectiveBinding<string>) {
    const { value } = binding
    if (!value) return

    const permissionStore = usePermissionStore()
    if (!permissionStore.permissionCodes.includes(value)) {
      el.parentNode?.removeChild(el)
    }
  },
}

export function setupPermissionDirective(app: App) {
  app.directive('permission', permissionDirective)
}
