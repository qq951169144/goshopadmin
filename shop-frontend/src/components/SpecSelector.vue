<template>
  <div class="spec-selector">
    <div v-for="spec in specifications" :key="spec.id" class="spec-row">
      <div class="spec-name">{{ spec.name }}</div>
      <div class="spec-values">
        <div
          v-for="value in spec.values"
          :key="value.id"
          class="spec-value"
          :class="{
            'active': isSelected(spec.id, value.id),
            'disabled': isDisabled(spec.id, value.id)
          }"
          @click="handleSelect(spec.id, value.id)"
        >
          <img v-if="value.image" :src="getImageUrl(value.image)" class="value-image" />
          <span>{{ value.value }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  specifications: {
    type: Array,
    default: () => []
  },
  selectedSpecs: {
    type: Object,
    default: () => ({})
  },
  skuList: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['change'])

// 获取完整的图片URL
const getImageUrl = (imageUrl) => {
  if (!imageUrl) return ''
  if (imageUrl.startsWith('http://') || imageUrl.startsWith('https://')) {
    return imageUrl
  }
  return `/api${imageUrl}`
}

// 判断是否已选择
const isSelected = (specId, valueId) => {
  // 使用宽松比较，因为类型可能不一致（字符串 vs 数字）
  return Number(props.selectedSpecs[specId]) === Number(valueId)
}

// 判断是否禁用（根据已选规格组合判断是否有可用SKU）
const isDisabled = (specId, valueId) => {
  // 如果已经选中了，一定可用
  if (isSelected(specId, valueId)) {
    return false
  }

  // 构建模拟选择状态：保留其他已选规格，当前规格使用传入的valueId
  const simulatedSpecs = { ...props.selectedSpecs }

  // 如果点击的是已选中的规格值，模拟取消选择
  if (simulatedSpecs[specId] === valueId) {
    delete simulatedSpecs[specId]
  } else {
    simulatedSpecs[specId] = valueId
  }

  // 检查是否有SKU包含这个规格组合
  const hasMatchingSKU = props.skuList.some(sku => {
    if (sku.status !== 'active') return false

    const specCombination = sku.spec_combination || {}

    // 检查是否匹配所有已选规格
    for (const [sId, vId] of Object.entries(simulatedSpecs)) {
      // 将两者都转为数字进行比较，因为后端返回的 spec_combination 中的 value 是数字
      // 而前端传入的 valueId 可能是字符串或数字
      const specValue = specCombination[sId]
      if (Number(specValue) !== Number(vId)) {
        return false
      }
    }
    return true
  })

  return !hasMatchingSKU
}

// 处理选择
const handleSelect = (specId, valueId) => {
  if (isDisabled(specId, valueId)) return

  const newSelectedSpecs = { ...props.selectedSpecs }

  // 如果已选择则取消选择，否则选择
  // 使用宽松比较，因为类型可能不一致（字符串 vs 数字）
  if (Number(newSelectedSpecs[specId]) === Number(valueId)) {
    delete newSelectedSpecs[specId]
  } else {
    newSelectedSpecs[specId] = valueId
  }

  emit('change', newSelectedSpecs)
}
</script>

<style scoped>
.spec-selector {
  padding: 10px 0;
}

.spec-row {
  margin-bottom: 16px;
}

.spec-row:last-child {
  margin-bottom: 0;
}

.spec-name {
  font-size: 14px;
  color: #666;
  margin-bottom: 10px;
  font-weight: 500;
}

.spec-values {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.spec-value {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
  color: #333;
  cursor: pointer;
  transition: all 0.3s ease;
  background-color: white;
}

.spec-value:hover:not(.disabled) {
  border-color: #ff4757;
}

.spec-value.active {
  border-color: #ff4757;
  background-color: #fff5f5;
  color: #ff4757;
}

.spec-value.disabled {
  opacity: 0.4;
  cursor: not-allowed;
  background-color: #f5f5f5;
}

.value-image {
  width: 24px;
  height: 24px;
  object-fit: cover;
  border-radius: 2px;
}
</style>
