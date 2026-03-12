<template>
  <div class="register">
    <h1>注册</h1>
    <form @submit.prevent="register">
      <div class="form-group">
        <label>用户名</label>
        <input type="text" v-model="username" placeholder="请输入用户名" required />
      </div>
      <div class="form-group">
        <label>密码</label>
        <input type="password" v-model="password" placeholder="请输入密码" required />
      </div>
      <div class="form-group">
        <label>确认密码</label>
        <input type="password" v-model="confirmPassword" placeholder="请确认密码" required />
      </div>
      <div class="form-group">
        <label>验证码</label>
        <div class="captcha">
          <input type="text" v-model="captcha" placeholder="请输入验证码" required />
          <img :src="captchaUrl" alt="验证码" @click="refreshCaptcha" />
        </div>
      </div>
      <button type="submit" :disabled="loading">{{ loading ? '注册中...' : '注册' }}</button>
      <div class="login-link">
        已有账号？<a href="/login">立即登录</a>
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
const confirmPassword = ref('')
const captcha = ref('')
const loading = ref(false)
const captchaUrl = ref('/api/captcha?' + new Date().getTime())

const refreshCaptcha = () => {
  captchaUrl.value = '/api/captcha?' + new Date().getTime()
  captcha.value = ''
}

const register = async () => {
  if (!username.value || !password.value || !confirmPassword.value || !captcha.value) {
    alert('请填写完整信息')
    return
  }
  
  if (password.value !== confirmPassword.value) {
    alert('两次输入的密码不一致')
    return
  }
  
  loading.value = true
  try {
    await authAPI.register({
      username: username.value,
      password: password.value,
      captcha: captcha.value,
      captcha_id: '1' // 实际项目中应该从验证码接口获取
    })
    
    alert('注册成功，请登录')
    router.push('/login')
  } catch (error) {
    console.error('注册失败:', error)
    alert('注册失败，请稍后重试')
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
.register {
  padding: 20px;
  max-width: 400px;
  margin: 0 auto;
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  margin-top: 50px;
}

h1 {
  margin-bottom: 20px;
  color: #333;
  text-align: center;
}

.form-group {
  margin-bottom: 15px;
}

label {
  display: block;
  margin-bottom: 5px;
  color: #666;
  font-size: 14px;
}

input {
  width: 100%;
  padding: 10px;
  border: 1px solid #ddd;
  border-radius: 4px;
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
  height: 40px;
  cursor: pointer;
  border-radius: 4px;
}

button {
  width: 100%;
  padding: 12px;
  background-color: #4CAF50;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  margin-top: 20px;
  font-size: 16px;
  transition: all 0.3s ease;
}

button:hover:not(:disabled) {
  background-color: #45a049;
}

button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.login-link {
  text-align: center;
  margin-top: 15px;
  font-size: 14px;
  color: #666;
}

.login-link a {
  color: #4CAF50;
  text-decoration: none;
}

.login-link a:hover {
  text-decoration: underline;
}
</style>