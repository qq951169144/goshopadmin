import { defineStore } from 'pinia'

export const useMessageStore = defineStore('message', {
    state: () => ({
        messages: [],
        unreadCount: 0
    }),

    actions: {
        addMessage(message) {
            const newMessage = {
                ...message,
                id: Date.now(),
                is_read: false,
                created_at: new Date().toISOString()
            }
            this.messages.unshift(newMessage)
            this.unreadCount++
            this.saveToLocal()
        },

        markAsRead(messageId) {
            const msg = this.messages.find(m => m.id === messageId)
            if (msg && !msg.is_read) {
                msg.is_read = true
                this.unreadCount = Math.max(0, this.unreadCount - 1)
                this.saveToLocal()
            }
        },

        markAllAsRead() {
            this.messages.forEach(msg => {
                msg.is_read = true
            })
            this.unreadCount = 0
            this.saveToLocal()
        },

        clearMessages() {
            this.messages = []
            this.unreadCount = 0
            this.saveToLocal()
        },

        saveToLocal() {
            localStorage.setItem('messages', JSON.stringify(this.messages))
            localStorage.setItem('unreadCount', this.unreadCount.toString())
        },

        loadFromLocal() {
            try {
                const messages = localStorage.getItem('messages')
                const unreadCount = localStorage.getItem('unreadCount')
                if (messages) {
                    this.messages = JSON.parse(messages)
                }
                if (unreadCount) {
                    this.unreadCount = parseInt(unreadCount)
                }
            } catch (error) {
                console.error('Failed to load messages from localStorage:', error)
            }
        }
    },

    getters: {
        unreadMessages: (state) => state.messages.filter(m => !m.is_read),
        readMessages: (state) => state.messages.filter(m => m.is_read),
        totalMessages: (state) => state.messages.length
    }
})
