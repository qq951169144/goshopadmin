<template>
  <div class="login">
    <div class="header">
      <button class="back-btn" @click="goHome">← 返回首页</button>
    </div>
    <h1>登录</h1>
    <form @submit.prevent="login">
      <div class="form-group">
        <label>用户名</label>
        <input type="text" v-model="username" placeholder="请输入用户名" required />
      </div>
      <div class="form-group">
        <label>密码</label>
        <input type="password" v-model="password" placeholder="请输入密码" required />
      </div>
      <div class="form-group">
        <label>验证码</label>
        <div class="captcha">
          <input type="text" v-model="captcha" placeholder="请输入验证码" required />
          <img :src="captchaUrl" alt="验证码" @click="refreshCaptcha" />
        </div>
      </div>
      <button type="submit" :disabled="loading">{{ loading ? '登录中...' : '登录' }}</button>
      <div class="register-link">
        还没有账号？<a href="/register">立即注册</a>
      </div>
    </form>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { authAPI } from '../api'

const router = useRouter()
const username = ref('')
const password = ref('')
const captcha = ref('')
const loading = ref(false)
const captchaUrl = ref('/api/captcha?' + new Date().getTime())

const refreshCaptcha = () => {
  captchaUrl.value = '/api/captcha?' + new Date().getTime()
  captcha.value = ''
}

const goHome = () => {
  router.push('/')
}

const login = async () => {
  if (!username.value || !password.value || !captcha.value) {
    alert('请填写完整信息')
    return
  }
  
  loading.value = true
  try {
    const response = await authAPI.login({
      username: username.value,
      password: password.value,
      captcha: captcha.value,
      captcha_id: '1' // 实际项目中应该从验证码接口获取
    })
    
    // 保存token
    localStorage.setItem('token', response.token)
    localStorage.setItem('customer_id', response.user_id)
    
    // 跳转到首页
    router.push('/')
  } catch (error) {
    console.error('登录失败:', error)
    alert('登录失败，请检查用户名、密码和验证码')
    refreshCaptcha()
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  refreshCaptcha()
})
</script>

<style scoped>
.login {
  padding: 20px;
  max-width: 400px;
  margin: 0 auto;
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  margin-top: 20px;
}

.header {
  margin-bottom: 20px;
}

.back-btn {
  background: none;
  border: none;
  color: #666;
  font-size: 14px;
  cursor: pointer;
  padding: 0;
  display: flex;
  align-items: center;
  gap: 4px;
}

.back-btn:hover {
  color: #4CAF50;
}

h1 {
  margin-bottom: 30px;
  color: #333;
  text-align: center;
}

.form-group {
  margin-bottom: 20px;
}

label {
  display: block;
  margin-bottom: 8px;
  color: #666;
  font-size: 14px;
}

input {
  width: 100%;
  padding: 12px;
  border: 1px solid #ddd;
  border-radius: 8px;
  font-size: 16px;
}

input:focus {
  outline: none;
  border-color: #4CAF50;
  box-shadow: 0 0 0 2px rgba(76, 175, 80, 0.1);
}

.captcha {
  display: flex;
  gap: 10px;
  align-items: center;
}

.captcha input {
  flex: 1;
}

.captcha img {
  width: 100px;
  height: 44px;
  cursor: pointer;
  border-radius: 8px;
}

button[type="submit"] {
  width: 100%;
  padding: 14px;
  background-color: #4CAF50;
  color: white;
  border: none;
  border-radius: 24px;
  cursor: pointer;
  margin-top: 20px;
  font-size: 16px;
  transition: all 0.3s ease;
}

button[type="submit"]:hover:not(:disabled) {
  background-color: #45a049;
}

button[type="submit"]:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.register-link {
  text-align: center;
  margin-top: 20px;
  font-size: 14px;
  color: #666;
}

.register-link a {
  color: #4CAF50;
  text-decoration: none;
}

.register-link a:hover {
  text-decoration: underline;
}
</style>
