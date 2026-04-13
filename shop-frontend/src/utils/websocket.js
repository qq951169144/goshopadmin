class WebSocketClient {
    constructor() {
        this.ws = null
        this.reconnectInterval = 5000
        this.listeners = new Map()
        this.isConnected = false
        this.shouldReconnect = true
        this.customerId = null
    }

    connect(customerId) {
        if (!customerId) {
            console.warn('WebSocket: No customer_id provided, skipping connection')
            return
        }

        this.customerId = customerId
        this.shouldReconnect = true

        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
        const wsUrl = `${protocol}//${window.location.host}/ws?customer_id=${customerId}`

        console.log('WebSocket connecting to:', wsUrl)

        try {
            this.ws = new WebSocket(wsUrl)

            this.ws.onopen = () => {
                console.log('WebSocket connected')
                this.isConnected = true
            }

            this.ws.onmessage = (event) => {
                try {
                    const message = JSON.parse(event.data)
                    console.log('WebSocket message received:', message)
                    this.emit(message.type, message.data)
                } catch (error) {
                    console.error('WebSocket message parse error:', error)
                }
            }

            this.ws.onclose = (event) => {
                console.log('WebSocket disconnected:', event.code, event.reason)
                this.isConnected = false

                if (this.shouldReconnect) {
                    console.log(`WebSocket reconnecting in ${this.reconnectInterval}ms...`)
                    setTimeout(() => this.reconnect(), this.reconnectInterval)
                }
            }

            this.ws.onerror = (error) => {
                console.error('WebSocket error:', error)
            }
        } catch (error) {
            console.error('WebSocket connection error:', error)
        }
    }

    on(type, callback) {
        if (!this.listeners.has(type)) {
            this.listeners.set(type, [])
        }
        this.listeners.get(type).push(callback)
    }

    emit(type, data) {
        const callbacks = this.listeners.get(type) || []
        callbacks.forEach(cb => {
            try {
                cb(data)
            } catch (error) {
                console.error('WebSocket listener error:', error)
            }
        })
    }

    off(type, callback) {
        if (!type) {
            this.listeners.clear()
            return
        }
        if (!callback) {
            this.listeners.delete(type)
            return
        }
        const callbacks = this.listeners.get(type) || []
        const index = callbacks.indexOf(callback)
        if (index > -1) {
            callbacks.splice(index, 1)
        }
    }

    disconnect() {
        this.shouldReconnect = false
        if (this.ws) {
            this.ws.close()
            this.ws = null
        }
        this.isConnected = false
        this.customerId = null
    }

    reconnect() {
        if (this.shouldReconnect && !this.isConnected && this.customerId) {
            this.connect(this.customerId)
        }
    }

    getConnectionStatus() {
        return this.isConnected
    }
}

export const wsClient = new WebSocketClient()
