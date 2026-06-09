<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '@/store/user'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const formRef = ref<FormInstance>()
const loading = ref(false)

const loginForm = reactive({
  shop_code: localStorage.getItem('shop_code') || '',
  username: '',
  password: '',
})

const rules = reactive<FormRules>({
  shop_code: [{ required: true, message: '请输入店铺编号', trigger: 'blur' }],
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
})

async function handleLogin() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    await userStore.login(loginForm)
    localStorage.setItem('shop_code', loginForm.shop_code)
    const redirect = (route.query.redirect as string) || '/shop/'
    router.push(redirect)
    ElMessage.success('登录成功')
  } catch (err: any) {
    ElMessage.error(err?.message || '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="shop-login">
    <div class="login-card">
      <h2 class="login-title">店铺管理系统</h2>
      <p class="login-subtitle">Shop Management System</p>
      <el-form
        ref="formRef"
        :model="loginForm"
        :rules="rules"
        label-position="top"
        class="login-form"
        @keyup.enter="handleLogin"
      >
        <el-form-item label="店铺编号" prop="shop_code">
          <el-input
            v-model="loginForm.shop_code"
            placeholder="请输入店铺编号"
            prefix-icon="Shop"
            size="large"
          />
        </el-form-item>
        <el-form-item label="用户名" prop="username">
          <el-input
            v-model="loginForm.username"
            placeholder="请输入用户名"
            prefix-icon="User"
            size="large"
          />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input
            v-model="loginForm.password"
            type="password"
            placeholder="请输入密码"
            prefix-icon="Lock"
            size="large"
            show-password
          />
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            size="large"
            class="login-btn"
            :loading="loading"
            @click="handleLogin"
          >
            登 录
          </el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<style scoped>
.shop-login {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #0f2027 0%, #203a43 50%, #2c5364 100%);
}

.login-card {
  width: 420px;
  padding: 40px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.2);
}

.login-title {
  text-align: center;
  font-size: 24px;
  font-weight: 600;
  color: #1a1a2e;
  margin: 0 0 4px;
}

.login-subtitle {
  text-align: center;
  font-size: 13px;
  color: #999;
  margin: 0 0 32px;
}

.login-form {
  margin-top: 8px;
}

.login-btn {
  width: 100%;
}
</style>
