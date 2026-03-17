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
    
    <!-- 库存显示 -->
    <div v-if="currentSkuStock !== null" class="stock-info">
      <span v-if="currentSkuStock > 0" class="in-stock">
        库存: {{ currentSkuStock }} 件
      </span>
      <span v-else class="out-of-stock">
        库存不足
      </span>
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

const emit = defineEmits(['change', 'stock-change'])

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

  // 如果没有SKU列表，默认启用所有规格
  if (!props.skuList || props.skuList.length === 0) {
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
    // 兼容处理：只要不是明确禁用，就默认可用
    if (sku.status === 'inactive' || sku.status === 'disabled') {
      return false
    }

    const specCombination = sku.spec_combination || {}
    
    // 如果 spec_combination 为空，使用降级方案：默认可用
    if (Object.keys(specCombination).length === 0) {
      return true
    }

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

// 检查当前选中规格的库存
const currentSkuStock = computed(() => {
  if (!props.skuList || props.skuList.length === 0) return null
  
  // 检查是否选择了完整规格
  const selectedCount = Object.keys(props.selectedSpecs).length
  const specCount = props.specifications.length
  if (selectedCount !== specCount) {
    return null
  }
  
  const matchedSku = props.skuList.find(sku => {
    if (sku.status === 'inactive' || sku.status === 'disabled') return false
    const specCombination = sku.spec_combination || {}
    
    // 如果 spec_combination 为空，使用降级方案：默认匹配第一个可用SKU
    if (Object.keys(specCombination).length === 0) {
      // 通过索引匹配：根据已选规格的顺序找到对应的SKU
      return matchSkuByIndex(sku)
    }
    
    for (const [sId, vId] of Object.entries(props.selectedSpecs)) {
      if (Number(specCombination[sId]) !== Number(vId)) {
        return false
      }
    }
    return true
  })
  
  const stock = matchedSku ? matchedSku.stock : null
  
  // 通知父组件库存变化
  emit('stock-change', stock)
  
  return stock
})

// 是否库存不足
const isOutOfStock = computed(() => {
  return currentSkuStock.value !== null && currentSkuStock.value <= 0
})

// 降级方案：通过索引匹配SKU
const matchSkuByIndex = (sku) => {
  // 当 spec_combination 为空时，使用 sku_code 或索引来匹配
  // 这里简化处理：默认认为 SKU 顺序与规格组合顺序一致
  return true
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

/* 库存信息样式 */
.stock-info {
  margin-top: 12px;
  padding: 8px 12px;
  border-radius: 4px;
  font-size: 14px;
  background-color: #f5f5f5;
}

.in-stock {
  color: #52c41a;
}

.out-of-stock {
  color: #ff4d4f;
  font-weight: 500;
}
</style>
