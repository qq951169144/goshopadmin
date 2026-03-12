import { defineStore } from 'pinia'
import api from '../api'

// 购物车存储
const useCartStore = defineStore('cart', {
  state: () => ({
    items: [],
    isLoading: false,
    error: null
  }),

  getters: {
    totalItems: (state) => state.items.reduce((total, item) => total + item.quantity, 0),
    totalPrice: (state) => state.items.reduce((total, item) => total + item.price * item.quantity, 0),
    cartId: () => localStorage.getItem('cart_id') || `local_${Date.now()}`
  },

  actions: {
    // 初始化购物车
    async initCart() {
      // 从LocalStorage加载购物车
      const savedCart = localStorage.getItem('cart')
      if (savedCart) {
        this.items = JSON.parse(savedCart)
      }

      // 如果用户已登录，同步到服务器
      const token = localStorage.getItem('token')
      if (token) {
        await this.syncCart()
      }
    },

    // 添加商品到购物车
    async addToCart(product) {
      this.isLoading = true
      this.error = null

      try {
        // 检查商品是否已在购物车中
        const existingItem = this.items.find(item => item.product_id === product.id)
        if (existingItem) {
          existingItem.quantity += 1
        } else {
          this.items.push({
            product_id: product.id,
            quantity: 1,
            price: product.price,
            sku: product.sku || 'default',
            name: product.name
          })
        }

        // 保存到LocalStorage
        this.saveToLocalStorage()

        // 如果用户已登录，同步到服务器
        const token = localStorage.getItem('token')
        if (token) {
          await api.post('/cart/items', {
            product_id: product.id,
            quantity: 1,
            price: product.price,
            sku: product.sku || 'default'
          })
        }
      } catch (error) {
        this.error = error.message
      } finally {
        this.isLoading = false
      }
    },

    // 更新商品数量
    async updateQuantity(itemId, quantity) {
      this.isLoading = true
      this.error = null

      try {
        const item = this.items.find(item => item.product_id === itemId)
        if (item) {
          item.quantity = quantity

          // 保存到LocalStorage
          this.saveToLocalStorage()

          // 如果用户已登录，同步到服务器
          const token = localStorage.getItem('token')
          if (token) {
            await api.put(`/cart/items/${itemId}`, {
              quantity: quantity
            })
          }
        }
      } catch (error) {
        this.error = error.message
      } finally {
        this.isLoading = false
      }
    },

    // 移除商品
    async removeItem(itemId) {
      this.isLoading = true
      this.error = null

      try {
        this.items = this.items.filter(item => item.product_id !== itemId)

        // 保存到LocalStorage
        this.saveToLocalStorage()

        // 如果用户已登录，同步到服务器
        const token = localStorage.getItem('token')
        if (token) {
          await api.delete(`/cart/items/${itemId}`)
        }
      } catch (error) {
        this.error = error.message
      } finally {
        this.isLoading = false
      }
    },

    // 清空购物车
    async clearCart() {
      this.isLoading = true
      this.error = null

      try {
        this.items = []

        // 保存到LocalStorage
        this.saveToLocalStorage()

        // 如果用户已登录，同步到服务器
        const token = localStorage.getItem('token')
        if (token) {
          // 调用清空购物车API
        }
      } catch (error) {
        this.error = error.message
      } finally {
        this.isLoading = false
      }
    },

    // 同步购物车到服务器
    async syncCart() {
      this.isLoading = true
      this.error = null

      try {
        const response = await api.post('/cart/sync', {
          items: this.items
        })

        // 更新购物车数据
        if (response.items) {
          this.items = response.items
          this.saveToLocalStorage()
        }
      } catch (error) {
        this.error = error.message
      } finally {
        this.isLoading = false
      }
    },

    // 从服务器获取购物车
    async fetchCart() {
      const token = localStorage.getItem('token')
      if (!token) return

      this.isLoading = true
      this.error = null

      try {
        const response = await api.get('/cart')
        if (response.items) {
          this.items = response.items
          this.saveToLocalStorage()
        }
      } catch (error) {
        this.error = error.message
      } finally {
        this.isLoading = false
      }
    },

    // 保存到LocalStorage
    saveToLocalStorage() {
      localStorage.setItem('cart', JSON.stringify(this.items))
      localStorage.setItem('cart_id', this.cartId)
    }
  }
})

export default useCartStore