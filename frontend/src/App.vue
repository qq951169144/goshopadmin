<template>
  <div id="app">
    <!-- 登录页面 -->
    <Login v-if="!isLoggedIn" />
    
    <!-- 主页 -->
    <Home v-else />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import Login from './components/Login.vue';
import Home from './components/Home.vue';

const isLoggedIn = ref(false);

// 检查登录状态
const checkLoginStatus = () => {
  const token = localStorage.getItem('token');
  isLoggedIn.value = !!token;
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
</style>