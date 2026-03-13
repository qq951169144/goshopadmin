<template>
  <div class="address-edit-page">
    <div class="page-header">
      <button class="back-btn" @click="goBack">←</button>
      <h1>{{ isEdit ? '编辑地址' : '新增地址' }}</h1>
    </div>

    <div class="form-content">
      <div class="form-group">
        <label>收货人</label>
        <input 
          type="text" 
          v-model="form.name" 
          placeholder="请输入收货人姓名"
        />
      </div>

      <div class="form-group">
        <label>手机号码</label>
        <input 
          type="tel" 
          v-model="form.phone" 
          placeholder="请输入手机号码"
          maxlength="11"
        />
      </div>

      <div class="form-group">
        <label>所在地区</label>
        <div class="region-select" @click="showRegionPicker = true">
          <span v-if="form.province && form.city && form.district">
            {{ form.province }} {{ form.city }} {{ form.district }}
          </span>
          <span v-else class="placeholder">请选择省/市/区</span>
          <span class="arrow">›</span>
        </div>
      </div>

      <div class="form-group">
        <label>详细地址</label>
        <textarea 
          v-model="form.detail_address" 
          placeholder="请输入街道、楼牌号等详细地址"
          rows="3"
        ></textarea>
      </div>

      <div class="form-group checkbox-group">
        <label class="checkbox-label">
          <input type="checkbox" v-model="form.is_default" />
          <span>设为默认地址</span>
        </label>
      </div>
    </div>

    <div class="bottom-bar">
      <button class="save-btn" @click="saveAddress">保存</button>
    </div>

    <!-- 地区选择器（简化版） -->
    <div v-if="showRegionPicker" class="region-picker-overlay" @click="showRegionPicker = false">
      <div class="region-picker" @click.stop>
        <div class="picker-header">
          <span>选择地区</span>
          <button @click="showRegionPicker = false">完成</button>
        </div>
        <div class="picker-content">
          <div class="picker-column">
            <div 
              v-for="province in provinces" 
              :key="province"
              class="picker-item"
              :class="{ active: form.province === province }"
              @click="selectProvince(province)"
            >
              {{ province }}
            </div>
          </div>
          <div class="picker-column">
            <div 
              v-for="city in cities" 
              :key="city"
              class="picker-item"
              :class="{ active: form.city === city }"
              @click="selectCity(city)"
            >
              {{ city }}
            </div>
          </div>
          <div class="picker-column">
            <div 
              v-for="district in districts" 
              :key="district"
              class="picker-item"
              :class="{ active: form.district === district }"
              @click="selectDistrict(district)"
            >
              {{ district }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { addressAPI } from '../api'

const route = useRoute()
const router = useRouter()

const form = ref({
  name: '',
  phone: '',
  province: '',
  city: '',
  district: '',
  detail_address: '',
  is_default: false
})

const isEdit = ref(false)
const addressId = ref(null)
const showRegionPicker = ref(false)
const fromCheckout = ref(false)

// 简化版的省市区数据
const regionData = {
  '广东省': {
    '深圳市': ['南山区', '福田区', '罗湖区', '宝安区', '龙岗区'],
    '广州市': ['天河区', '越秀区', '海珠区', '白云区', '番禺区'],
    '东莞市': ['南城街道', '东城街道', '莞城街道', '万江街道']
  },
  '北京市': {
    '北京市': ['朝阳区', '海淀区', '东城区', '西城区', '丰台区']
  },
  '上海市': {
    '上海市': ['浦东新区', '黄浦区', '徐汇区', '长宁区', '静安区']
  }
}

const provinces = computed(() => Object.keys(regionData))

const cities = computed(() => {
  if (!form.value.province) return []
  return Object.keys(regionData[form.value.province] || {})
})

const districts = computed(() => {
  if (!form.value.province || !form.value.city) return []
  return regionData[form.value.province]?.[form.value.city] || []
})

onMounted(() => {
  addressId.value = route.params.id
  isEdit.value = !!addressId.value
  fromCheckout.value = route.query.from === 'checkout'
  
  if (isEdit.value) {
    loadAddressDetail()
  }
})

const loadAddressDetail = async () => {
  try {
    const response = await addressAPI.getAddress(addressId.value)
    if (response) {
      form.value = { ...response }
    }
  } catch (error) {
    console.error('加载地址详情失败:', error)
  }
}

const selectProvince = (province) => {
  form.value.province = province
  form.value.city = ''
  form.value.district = ''
}

const selectCity = (city) => {
  form.value.city = city
  form.value.district = ''
}

const selectDistrict = (district) => {
  form.value.district = district
}

const validateForm = () => {
  if (!form.value.name.trim()) {
    alert('请输入收货人姓名')
    return false
  }
  if (!form.value.phone.trim()) {
    alert('请输入手机号码')
    return false
  }
  if (!/^1[3-9]\d{9}$/.test(form.value.phone)) {
    alert('请输入正确的手机号码')
    return false
  }
  if (!form.value.province || !form.value.city || !form.value.district) {
    alert('请选择所在地区')
    return false
  }
  if (!form.value.detail_address.trim()) {
    alert('请输入详细地址')
    return false
  }
  return true
}

const saveAddress = async () => {
  if (!validateForm()) return

  try {
    if (isEdit.value) {
      await addressAPI.updateAddress(addressId.value, form.value)
      alert('地址更新成功')
    } else {
      await addressAPI.createAddress(form.value)
      alert('地址添加成功')
    }
    
    if (fromCheckout.value) {
      router.push('/checkout')
    } else {
      router.push('/addresses')
    }
  } catch (error) {
    console.error('保存地址失败:', error)
    alert('保存失败，请重试')
  }
}

const goBack = () => {
  router.back()
}
</script>

<style scoped>
.address-edit-page {
  min-height: 100vh;
  background-color: #f5f5f5;
  padding-bottom: 80px;
}

.page-header {
  display: flex;
  align-items: center;
  background-color: white;
  padding: 12px 16px;
  border-bottom: 1px solid #eee;
}

.back-btn {
  background: none;
  border: none;
  font-size: 20px;
  color: #333;
  cursor: pointer;
  padding: 4px 8px;
  margin-right: 12px;
}

.page-header h1 {
  font-size: 18px;
  color: #333;
  margin: 0;
  flex: 1;
}

/* 表单内容 */
.form-content {
  background-color: white;
  margin-top: 12px;
}

.form-group {
  padding: 16px;
  border-bottom: 1px solid #f5f5f5;
}

.form-group label {
  display: block;
  font-size: 14px;
  color: #333;
  margin-bottom: 8px;
}

.form-group input,
.form-group textarea {
  width: 100%;
  border: none;
  outline: none;
  font-size: 15px;
  color: #333;
  background: none;
}

.form-group input::placeholder,
.form-group textarea::placeholder {
  color: #999;
}

.form-group textarea {
  resize: none;
}

.region-select {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 15px;
  color: #333;
  cursor: pointer;
}

.region-select .placeholder {
  color: #999;
}

.region-select .arrow {
  color: #999;
  font-size: 18px;
}

.checkbox-group {
  display: flex;
  align-items: center;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  margin-bottom: 0;
}

.checkbox-label input[type="checkbox"] {
  width: 18px;
  height: 18px;
  accent-color: #4CAF50;
}

/* 底部保存按钮 */
.bottom-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background-color: white;
  padding: 12px 16px;
  border-top: 1px solid #eee;
  box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.05);
}

.save-btn {
  width: 100%;
  padding: 14px;
  background-color: #4CAF50;
  color: white;
  border: none;
  border-radius: 24px;
  font-size: 16px;
  cursor: pointer;
}

.save-btn:hover {
  background-color: #45a049;
}

/* 地区选择器 */
.region-picker-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  z-index: 1000;
  display: flex;
  align-items: flex-end;
}

.region-picker {
  width: 100%;
  background-color: white;
  border-radius: 16px 16px 0 0;
  overflow: hidden;
}

.picker-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  border-bottom: 1px solid #eee;
}

.picker-header span {
  font-size: 16px;
  font-weight: bold;
}

.picker-header button {
  background: none;
  border: none;
  color: #4CAF50;
  font-size: 14px;
  cursor: pointer;
}

.picker-content {
  display: flex;
  height: 300px;
  overflow: hidden;
}

.picker-column {
  flex: 1;
  overflow-y: auto;
  text-align: center;
}

.picker-item {
  padding: 12px;
  font-size: 14px;
  color: #333;
  cursor: pointer;
}

.picker-item.active {
  color: #4CAF50;
  font-weight: bold;
}
</style>
