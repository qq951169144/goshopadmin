class WebSocketClient {
    constructor() {
        this.ws = null
        this.reconnectInterval = 5000
        this.listeners = new Map()
        this.isConnected = false
        this.shouldReconnect = true
    }

    connect() {
        const token = localStorage.getItem('token')
        if (!token) {
            console.warn('[WS] 未提供 token，跳过连接')
            return
        }

        this.shouldReconnect = true

        const wsUrl = `ws://localhost:8081/ws?token=${encodeURIComponent(token)}`

        console.log(`[WS] 正在连接 | URL: ${wsUrl}`)

        try {
            this.ws = new WebSocket(wsUrl, [token])

            this.ws.onopen = () => {
                console.log('[WS] 连接成功')
                this.isConnected = true
            }

            this.ws.onmessage = (event) => {
                try {
                    const message = JSON.parse(event.data)
                    console.log(`[WS] 收到原始消息 | JSON: %o`, message)

                    const messageType = message.type
                    const messageData = message.data

                    switch (messageType) {
                        case 'order_created':
                            console.log('[WS] 收到消息 | 类型: order_created (订单创建) | 数据:', messageData)
                            break
                        case 'order_paid':
                            console.log('[WS] 收到消息 | 类型: order_paid (订单已支付) | 数据:', messageData)
                            break
                        case 'order_shipped':
                            console.log('[WS] 收到消息 | 类型: order_shipped (订单已发货) | 数据:', messageData)
                            break
                        case 'order_received':
                            console.log('[WS] 收到消息 | 类型: order_received (订单已收货) | 数据:', messageData)
                            break
                        case 'order_canceled':
                            console.log('[WS] 收到消息 | 类型: order_canceled (订单已取消) | 数据:', messageData)
                            break
                        case 'system_notice':
                            console.log('[WS] 收到消息 | 类型: system_notice (系统通知) | 数据:', messageData)
                            break
                        default:
                            console.log(`[WS] 收到未知类型消息 | 类型: ${messageType} | 数据:`, messageData)
                    }

                    this.emit(messageType, messageData)
                } catch (error) {
                    console.error('[WS] 消息解析失败 | 错误:', error)
                }
            }

            this.ws.onclose = (event) => {
                console.log(`[WS] 连接断开 | code: ${event.code} | reason: ${event.reason || '无'}`)
                this.isConnected = false

                if (this.shouldReconnect) {
                    console.log(`[WS] 将在 ${this.reconnectInterval}ms 后自动重连...`)
                    setTimeout(() => this.reconnect(), this.reconnectInterval)
                }
            }

            this.ws.onerror = (error) => {
                console.error('[WS] 连接错误 | 错误:', error)
            }
        } catch (error) {
            console.error('[WS] 连接异常 | 错误:', error)
        }
    }

    on(type, callback) {
        if (!this.listeners.has(type)) {
            this.listeners.set(type, [])
        }
        console.log(`[WS] 注册监听器 | 类型: ${type}`)
        this.listeners.get(type).push(callback)
    }

    emit(type, data) {
        const callbacks = this.listeners.get(type) || []
        console.log(`[WS] 触发监听器 | 类型: ${type} | 回调数量: ${callbacks.length}`)
        callbacks.forEach(cb => {
            try {
                cb(data)
            } catch (error) {
                console.error(`[WS] 监听器执行错误 | 类型: ${type} | 错误:`, error)
            }
        })
    }

    off(type, callback) {
        if (!type) {
            this.listeners.clear()
            console.log('[WS] 已清除所有监听器')
            return
        }
        if (!callback) {
            this.listeners.delete(type)
            console.log(`[WS] 已移除监听器 | 类型: ${type}`)
            return
        }
        const callbacks = this.listeners.get(type) || []
        const index = callbacks.indexOf(callback)
        if (index > -1) {
            callbacks.splice(index, 1)
            console.log(`[WS] 已移除指定监听器 | 类型: ${type}`)
        }
    }

    disconnect() {
        console.log('[WS] 主动断开连接')
        this.shouldReconnect = false
        if (this.ws) {
            this.ws.close()
            this.ws = null
        }
        this.isConnected = false
    }

    reconnect() {
        if (this.shouldReconnect && !this.isConnected) {
            console.log('[WS] 执行重连')
            this.connect()
        }
    }

    getConnectionStatus() {
        return this.isConnected
    }
}

export const wsClient = new WebSocketClient()
