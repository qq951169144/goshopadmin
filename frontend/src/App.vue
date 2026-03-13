<template>
  <div id="app">
    <!-- 登录页面 -->
    <Login v-if="!isLoggedIn && !checkingLogin" />
    
    <!-- 主页 -->
    <Home v-else-if="isLoggedIn" />
    
    <!-- 加载中 -->
    <div v-else class="loading-container">
      <div class="loading-spinner"></div>
      <p>加载中...</p>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import Login from './views/Login.vue';
import Home from './views/Home.vue';

const isLoggedIn = ref(false);
const checkingLogin = ref(true);

// 检查登录状态
const checkLoginStatus = () => {
  setTimeout(() => {
    const token = localStorage.getItem('token');
    isLoggedIn.value = !!token;
    checkingLogin.value = false;
  }, 100);
};

// 初始化
onMounted(() => {
  checkLoginStatus();
});
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: Arial, sans-serif;
  background-color: #f5f7fa;
}

#app {
  height: 100vh;
  width: 100%;
}

.loading-container {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.loading-spinner {
  width: 50px;
  height: 50px;
  border: 5px solid rgba(255, 255, 255, 0.3);
  border-radius: 50%;
  border-top-color: white;
  animation: spin 1s ease-in-out infinite;
  margin-bottom: 20px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.loading-container p {
  font-size: 18px;
  font-weight: 500;
}
</style>