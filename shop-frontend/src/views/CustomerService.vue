<template>
  <div class="customer-service-page">
    <div class="page-header">
      <button class="back-btn" @click="goBack">←</button>
      <h1>客服中心</h1>
    </div>

    <div class="service-content">
      <div class="service-intro">
        <div class="service-icon">💬</div>
        <h2>在线客服</h2>
        <p>工作时间：周一至周日 9:00 - 21:00</p>
      </div>

      <div class="service-options">
        <div class="service-item" @click="contactService">
          <div class="item-icon">📱</div>
          <div class="item-info">
            <h3>在线咨询</h3>
            <p>点击开始与客服对话</p>
          </div>
          <span class="arrow">→</span>
        </div>

        <div class="service-item" @click="callService">
          <div class="item-icon">📞</div>
          <div class="item-info">
            <h3>电话客服</h3>
            <p>400-888-8888</p>
          </div>
          <span class="arrow">→</span>
        </div>

        <div class="service-item" @click="viewFAQ">
          <div class="item-icon">❓</div>
          <div class="item-info">
            <h3>常见问题</h3>
            <p>查看FAQ</p>
          </div>
          <span class="arrow">→</span>
        </div>
      </div>

      <div class="service-tips">
        <h3>温馨提示</h3>
        <ul>
          <li>活动订单相关问题请通过活动订单详情页联系我们</li>
          <li>如有紧急问题，建议电话联系客服</li>
          <li>我们将在24小时内回复您的邮件</li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'

const router = useRouter()

const goBack = () => {
  router.back()
}

const contactService = () => {
  alert('在线客服功能开发中，请拨打客服电话或通过活动订单页联系我们')
}

const callService = () => {
  window.location.href = 'tel:400-888-8888'
}

const viewFAQ = () => {
  alert('FAQ功能开发中')
}
</script>

<style scoped>
.customer-service-page {
  min-height: 100vh;
  background-color: #f5f5f5;
}

.page-header {
  background-color: #fff;
  padding: 12px 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid #eee;
  position: sticky;
  top: 0;
  z-index: 10;
}

.back-btn {
  position: absolute;
  left: 16px;
  background: none;
  border: none;
  font-size: 20px;
  color: #333;
  padding: 4px 8px;
  cursor: pointer;
}

.page-header h1 {
  font-size: 16px;
  color: #333;
  margin: 0;
}

.service-content {
  padding: 16px;
}

.service-intro {
  background-color: #fff;
  border-radius: 12px;
  padding: 24px;
  text-align: center;
  margin-bottom: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.service-icon {
  font-size: 48px;
  margin-bottom: 12px;
}

.service-intro h2 {
  font-size: 18px;
  color: #333;
  margin: 0 0 8px 0;
}

.service-intro p {
  font-size: 13px;
  color: #999;
  margin: 0;
}

.service-options {
  background-color: #fff;
  border-radius: 12px;
  margin-bottom: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.service-item {
  display: flex;
  align-items: center;
  padding: 16px;
  border-bottom: 1px solid #f5f5f5;
  cursor: pointer;
  transition: background-color 0.3s ease;
}

.service-item:last-child {
  border-bottom: none;
}

.service-item:active {
  background-color: #f5f5f5;
}

.item-icon {
  width: 40px;
  height: 40px;
  background-color: #f5f5f5;
  border-radius: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  margin-right: 12px;
}

.item-info {
  flex: 1;
}

.item-info h3 {
  font-size: 14px;
  color: #333;
  margin: 0 0 4px 0;
}

.item-info p {
  font-size: 12px;
  color: #999;
  margin: 0;
}

.arrow {
  color: #ccc;
  font-size: 16px;
}

.service-tips {
  background-color: #fff;
  border-radius: 12px;
  padding: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.service-tips h3 {
  font-size: 14px;
  color: #333;
  margin: 0 0 12px 0;
}

.service-tips ul {
  margin: 0;
  padding-left: 20px;
}

.service-tips li {
  font-size: 13px;
  color: #666;
  line-height: 1.8;
}
</style>
