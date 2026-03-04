<template>
  <div class="login-container">
    <div class="login-form">
      <h2>商城后台管理系统 - 2.0</h2>
      <el-form :model="loginForm" :rules="rules" ref="loginFormRef" label-width="80px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="loginForm.username" placeholder="请输入用户名" />
        </el-form-item>
        
        <el-form-item label="密码" prop="password">
          <el-input v-model="loginForm.password" type="password" placeholder="请输入密码" />
        </el-form-item>
        
        <el-form-item label="验证码" prop="captcha">
          <div class="slider-captcha">
            <div class="captcha-image-container">
              <img :src="captchaUrl" class="captcha-image" ref="captchaImage" />
              <!-- 拼图块 -->
              <div class="puzzle-block" :style="{ left: sliderPosition + 'px', top: '50%', transform: 'translateY(-50%)' }"></div>
            </div>
            <div class="slider-track" ref="sliderTrack" 
                 @mousedown="startDrag" @touchstart="startDrag"
                 @mousemove="onDrag" @mouseup="stopDrag" @mouseleave="stopDrag"
                 @touchmove="onDrag" @touchend="stopDrag">
              <div class="slider-progress" :style="{ width: sliderPosition + 'px' }"></div>
              <div class="slider-block" :style="{ left: sliderPosition + 'px' }">
                <div class="slider-icon" :class="{ 'success': captchaHint === '验证成功' }">&rarr;</div>
              </div>
            </div>
            <div class="captcha-hint" :class="{ 'success': captchaHint === '验证成功' }">{{ captchaHint }}</div>
          </div>
        </el-form-item>
        
        <el-form-item>
          <el-button type="primary" @click="handleLogin" class="login-button">登录</el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted } from 'vue';
import { ElMessage } from 'element-plus';
import { authApi } from '../api/auth';

const loginFormRef = ref(null);
const captchaUrl = ref('');
const sliderTrack = ref(null);
const captchaImage = ref(null);
const sliderPosition = ref(0);
const isDragging = ref(false);
const captchaHint = ref('请拖动滑块完成验证');
const loginForm = reactive({
  username: '',
  password: '',
  captcha: '',
  captchaId: ''
});
// 验证码图片原始宽度（与后端一致）
const CAPTCHA_WIDTH = 300;

const rules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' }
  ],
  captcha: [
    { required: true, message: '请完成滑块验证', trigger: 'change' }
  ]
};

// 刷新验证码
const refreshCaptcha = async () => {
  try {
    const response = await authApi.getCaptcha();
    if (response.code === 200) {
      // 后端返回的是base64格式的图片
      captchaUrl.value = `data:image/png;base64,${response.data.image}`;
      // 保存captchaId
      loginForm.captchaId = response.data.id;
      // 重置滑块位置
      sliderPosition.value = 0;
      captchaHint.value = '请拖动滑块完成验证';
      loginForm.captcha = '';
      // 移除成功状态样式
      const sliderBlock = document.querySelector('.slider-block');
      if (sliderBlock) {
        sliderBlock.classList.remove('success');
      }
    }
  } catch (error) {
    ElMessage.error('获取验证码失败');
  }
};

// 开始拖动
const startDrag = (e) => {
  isDragging.value = true;
  // 阻止默认行为
  if (e.type === 'mousedown') {
    e.preventDefault();
  }
};

// 拖动中
const onDrag = (e) => {
  if (!isDragging.value) return;
  
  let clientX;
  if (e.type === 'mousemove') {
    clientX = e.clientX;
  } else if (e.type === 'touchmove') {
    e.preventDefault();
    clientX = e.touches[0].clientX;
  }
  
  if (sliderTrack.value) {
    const rect = sliderTrack.value.getBoundingClientRect();
    let newPosition = clientX - rect.left;
    
    // 限制滑块范围
    if (newPosition < 0) {
      newPosition = 0;
    } else if (newPosition > rect.width) {
      newPosition = rect.width;
    }
    
    sliderPosition.value = newPosition;
    
    // 根据验证码图片宽度计算实际答案
    if (captchaImage.value) {
      const imageWidth = captchaImage.value.width || CAPTCHA_WIDTH;
      // 计算比例并转换为基于300px宽度的答案
      const scale = CAPTCHA_WIDTH / imageWidth;
      const actualAnswer = Math.round(newPosition * scale);
      loginForm.captcha = actualAnswer;
    } else {
      //  fallback: 直接使用滑块位置
      loginForm.captcha = Math.round(newPosition);
    }
  }
};

// 停止拖动
const stopDrag = () => {
  if (isDragging.value) {
    isDragging.value = false;
    
    // 只更新滑块位置，不验证验证码
    // 验证码将在登录时验证
    if (loginForm.captcha && loginForm.captchaId) {
      captchaHint.value = '请点击登录按钮';
      // 确保滑块与进度条末端重合
      if (sliderTrack.value) {
        const rect = sliderTrack.value.getBoundingClientRect();
        // 计算滑块应该移动的位置，使滑块与进度条末端重合
        const blockWidth = 60; // 滑块宽度
        const newPosition = rect.width - blockWidth;
        sliderPosition.value = newPosition;
        // 注意：这里不修改 loginForm.captcha，保持当前值
      }
    }
  }
};

// 处理登录
const handleLogin = async () => {
  if (!loginFormRef.value) return;
  
  // 检查验证码是否已验证
  if (!loginForm.captcha) {
    ElMessage.error('请完成滑块验证');
    return;
  }
  
  await loginFormRef.value.validate(async (valid) => {
    if (valid) {
      try {
        // 验证码已经在停止拖动时验证过，这里直接登录
        const loginResponse = await authApi.login({
          username: loginForm.username,
          password: loginForm.password,
          captcha_id: loginForm.captchaId,
          captcha_ans: parseInt(loginForm.captcha)
        });
        
        if (loginResponse.code === 200) {
          // 保存 token 和用户信息
          localStorage.setItem('token', loginResponse.data.token);
          localStorage.setItem('user', JSON.stringify(loginResponse.data.user));
          
          ElMessage.success('登录成功');
          // 跳转到首页
          window.location.href = '/';
        } else {
          ElMessage.error(loginResponse.message || '登录失败');
          refreshCaptcha();
        }
      } catch (error) {
        ElMessage.error('登录失败，请重试');
        refreshCaptcha();
      }
    }
  });
};

// 全局鼠标事件处理
const handleGlobalMouseMove = (e) => {
  onDrag(e);
};

const handleGlobalMouseUp = () => {
  stopDrag();
};

// 全局触摸事件处理
const handleGlobalTouchMove = (e) => {
  onDrag(e);
};

const handleGlobalTouchEnd = () => {
  stopDrag();
};

// 初始化
onMounted(() => {
  refreshCaptcha();
  // 添加全局事件监听器
  window.addEventListener('mousemove', handleGlobalMouseMove);
  window.addEventListener('mouseup', handleGlobalMouseUp);
  window.addEventListener('touchmove', handleGlobalTouchMove, { passive: false });
  window.addEventListener('touchend', handleGlobalTouchEnd);
});

// 清理
onUnmounted(() => {
  // 移除全局事件监听器
  window.removeEventListener('mousemove', handleGlobalMouseMove);
  window.removeEventListener('mouseup', handleGlobalMouseUp);
  window.removeEventListener('touchmove', handleGlobalTouchMove);
  window.removeEventListener('touchend', handleGlobalTouchEnd);
});
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.login-form {
  background: white;
  padding: 40px;
  border-radius: 10px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.1);
  width: 100%;
  max-width: 400px;
}

.login-form h2 {
  text-align: center;
  margin-bottom: 30px;
  color: #333;
}

.slider-captcha {
  width: 100%;
}

.captcha-image-container {
  position: relative;
  width: 100%;
  height: 150px;
  margin-bottom: 10px;
  border-radius: 4px;
  overflow: hidden;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
}

.captcha-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.puzzle-block {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  width: 50px;
  height: 50px;
  background: rgba(255, 255, 255, 0.9);
  border: 2px solid #ddd;
  border-radius: 50%;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.2);
  z-index: 5;
}

.slider-track {
  position: relative;
  width: 100%;
  height: 40px;
  background: #f0f0f0;
  border-radius: 20px;
  overflow: hidden;
  cursor: pointer;
}

.slider-progress {
  position: absolute;
  top: 0;
  left: 0;
  height: 100%;
  background: linear-gradient(90deg, #667eea, #764ba2);
  border-radius: 20px;
  transition: width 0.1s ease;
}

.slider-block {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  width: 60px;
  height: 36px;
  background: #1890ff;
  border: none;
  border-radius: 18px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.2);
  cursor: pointer;
  display: flex;
  justify-content: center;
  align-items: center;
  user-select: none;
  z-index: 10;
  transition: all 0.3s ease;
}

.slider-icon {
  font-size: 16px;
  color: white;
  transition: all 0.3s ease;
}

.slider-block.success {
  background: #52c41a;
}

.slider-icon.success {
  color: white;
}

.captcha-hint {
  margin-top: 10px;
  font-size: 12px;
  color: #666;
  text-align: center;
  transition: all 0.3s ease;
}

.captcha-hint.success {
  color: #4CAF50;
  font-weight: bold;
}

.login-button {
  width: 100%;
  padding: 12px;
  font-size: 16px;
}
</style>