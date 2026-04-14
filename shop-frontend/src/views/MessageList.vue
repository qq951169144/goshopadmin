<template>
  <div class="message-list-page">
    <div class="page-header">
      <button class="back-btn" @click="goBack">←</button>
      <h1>站内信</h1>
      <button v-if="messages.length > 0" class="clear-btn" @click="markAllAsRead">全部已读</button>
    </div>

    <div class="message-list">
      <div v-if="messages.length === 0" class="empty-message">
        <div class="empty-icon">📭</div>
        <p>暂无站内信</p>
      </div>
      <div v-else>
        <div v-for="msg in messages" :key="msg.id"
          class="message-card"
          :class="{ unread: !msg.is_read }"
          @click="viewMessage(msg)">
          <div class="message-icon">{{ getMessageIcon(msg.type) }}</div>
          <div class="message-content">
            <div class="message-header">
              <div class="message-title">{{ getMessageTitle(msg.type) }}</div>
              <div class="message-time">{{ formatTime(msg.created_at) }}</div>
            </div>
            <div class="message-desc">{{ getMessageDesc(msg) }}</div>
          </div>
          <div v-if="!msg.is_read" class="unread-dot"></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useMessageStore } from '../store/message'

const router = useRouter()
const messageStore = useMessageStore()

const messages = computed(() => messageStore.messages)

const messageTypes = [
  'order_created',
  'order_paid',
  'order_shipped',
  'order_received',
  'order_canceled',
  'system_notice'
]

onMounted(() => {
  messageStore.loadFromLocal()
})

onUnmounted(() => {
})

const messageConfig = {
  order_created: { icon: '⚡', title: '订单已提交', desc: '您的活动订单已提交，正在处理中' },
  order_paid: { icon: '💰', title: '订单已支付', desc: '您的订单已支付成功' },
  order_shipped: { icon: '🚚', title: '商品已发货', desc: '您的商品已发货，请注意查收' },
  order_received: { icon: '✅', title: '订单已完成', desc: '感谢您的购买，欢迎再次光临' },
  order_canceled: { icon: '❌', title: '订单已取消', desc: '您的订单已取消' },
  system_notice: { icon: '📢', title: '系统通知', desc: '' }
}

const getMessageIcon = (type) => {
  return messageConfig[type]?.icon || '📌'
}

const getMessageTitle = (type) => {
  return messageConfig[type]?.title || '系统消息'
}

const getMessageDesc = (msg) => {
  if (msg.data) {
    if (msg.type === 'order_canceled') {
      return `订单号：${msg.data.order_no || ''}，${msg.data.reason || '已取消'}`
    }
    if (msg.type === 'order_created') {
      return `订单号：${msg.data.order_no || ''}，金额：¥${msg.data.total_amount || '0.00'}`
    }
    if (msg.type === 'order_paid' || msg.type === 'order_shipped' || msg.type === 'order_received') {
      return `订单号：${msg.data.order_no || ''}`
    }
  }
  return messageConfig[msg.type]?.desc || ''
}

const formatTime = (time) => {
  if (!time) return ''
  const date = new Date(time)
  const now = new Date()
  const diff = now - date
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)

  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes}分钟前`
  if (hours < 24) return `${hours}小时前`
  if (days < 7) return `${days}天前`

  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

const goBack = () => {
  router.back()
}

const viewMessage = (msg) => {
  messageStore.markAsRead(msg.id)
  if (msg.type.startsWith('order_') && msg.data?.order_id) {
    router.push(`/activity/order/${msg.data.order_id}`)
  }
}

const markAllAsRead = () => {
  messageStore.markAllAsRead()
}
</script>

<style scoped>
.message-list-page {
  min-height: 100vh;
  background-color: #f5f5f5;
}

.page-header {
  background-color: #fff;
  padding: 12px 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid #eee;
  position: sticky;
  top: 0;
  z-index: 10;
}

.back-btn {
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
  flex: 1;
  text-align: center;
}

.clear-btn {
  background: none;
  border: none;
  font-size: 13px;
  color: #4CAF50;
  cursor: pointer;
  padding: 4px 8px;
}

.message-list {
  padding: 12px;
}

.empty-message {
  text-align: center;
  padding: 60px 20px;
  color: #999;
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.message-card {
  background-color: #fff;
  border-radius: 12px;
  margin-bottom: 12px;
  padding: 16px;
  display: flex;
  gap: 12px;
  position: relative;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  cursor: pointer;
  transition: all 0.3s ease;
}

.message-card:active {
  background-color: #f5f5f5;
}

.message-card.unread {
  background-color: #fff;
}

.message-card.unread .message-title {
  font-weight: bold;
}

.message-icon {
  width: 40px;
  height: 40px;
  background-color: #f5f5f5;
  border-radius: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  flex-shrink: 0;
}

.message-content {
  flex: 1;
  min-width: 0;
}

.message-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
}

.message-title {
  font-size: 14px;
  color: #333;
}

.message-time {
  font-size: 12px;
  color: #999;
}

.message-desc {
  font-size: 13px;
  color: #666;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.unread-dot {
  position: absolute;
  top: 16px;
  right: 16px;
  width: 8px;
  height: 8px;
  background-color: #ff4757;
  border-radius: 50%;
}
</style>
