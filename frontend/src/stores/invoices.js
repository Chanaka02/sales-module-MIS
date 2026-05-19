import { defineStore } from 'pinia'
import api from '../services/api'

export const useInvoicesStore = defineStore('invoices', {
  state: () => ({
    invoices: [],
    currentInvoice: null,
    pagination: { total: 0, page: 1, limit: 20, totalPages: 0 },
    loading: false
  }),
  actions: {
    async fetchInvoices(page = 1, filters = {}) {
      this.loading = true
      try {
        const params = { page, limit: 20, ...filters }
        Object.keys(params).forEach(k => { if (!params[k]) delete params[k] })
        const { data } = await api.get('/invoices', { params })
        this.invoices = data.data || []
        this.pagination = { total: data.total, page: data.page, limit: data.limit, totalPages: data.total_pages }
      } finally { this.loading = false }
    },
    async fetchInvoice(id) {
      this.loading = true
      try {
        const { data } = await api.get(`/invoices/${id}`)
        this.currentInvoice = data
        return data
      } finally { this.loading = false }
    },
    async generateInvoice(saleId) {
      const { data } = await api.post(`/invoices/generate/${saleId}`)
      return data
    },
    async deleteInvoice(id) {
      await api.delete(`/invoices/${id}`)
      this.invoices = this.invoices.filter(i => i.id !== id)
    },
    async updateStatus(id, status) {
      await api.patch(`/invoices/${id}/status`, { status })
      const inv = this.invoices.find(i => i.id === id)
      if (inv) inv.status = status
    }
  }
})
