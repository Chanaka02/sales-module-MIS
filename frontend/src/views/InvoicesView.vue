<template>
  <div class="invoices-page">
    <div class="page-header">
      <div>
        <h1>Invoices</h1>
        <p>View and manage generated invoices</p>
      </div>
      <button class="btn btn-primary" @click="openGenerateModal">Generate Invoice</button>
    </div>

    <div class="filters-bar">
      <select v-model="statusFilter" @change="loadInvoices" class="form-control">
        <option value="">All Status</option>
        <option value="draft">Draft</option>
        <option value="sent">Sent</option>
        <option value="paid">Paid</option>
        <option value="overdue">Overdue</option>
      </select>
    </div>

    <LoadingSpinner v-if="store.loading" />

    <template v-else>
      <div class="card">
        <table class="data-table">
          <thead>
            <tr>
              <th>Invoice #</th>
              <th>Customer</th>
              <th>Issue Date</th>
              <th>Due Date</th>
              <th>Subtotal</th>
              <th>Discount</th>
              <th>Total</th>
              <th>Status</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="inv in store.invoices" :key="inv.id">
              <td class="text-primary font-medium">{{ inv.invoice_number }}</td>
              <td>{{ inv.customer_name || 'N/A' }}</td>
              <td class="text-muted">{{ inv.issue_date }}</td>
              <td class="text-muted">{{ inv.due_date || '-' }}</td>
              <td>Rs. {{ fmt(inv.subtotal) }}</td>
              <td class="text-warning">Rs. {{ fmt(inv.discount) }}</td>
              <td class="text-primary font-semibold">Rs. {{ fmt(inv.total) }}</td>
              <td><StatusBadge :status="inv.status" /></td>
              <td>
                <button class="btn btn-ghost btn-sm" @click="viewInvoice(inv.id)" title="View"><i class='bx bx-show'></i></button>
                <button v-if="inv.status !== 'paid'" class="btn btn-success btn-sm" @click="updateStatus(inv.id, 'paid')" title="Mark Paid"><i class='bx bx-check'></i></button>
                <button class="btn btn-danger btn-sm" style="background:#fee2e2; color:#ef4444; border:none;" @click="deleteInvoice(inv.id)" title="Delete"><i class='bx bx-trash'></i></button>
              </td>
            </tr>
            <tr v-if="!store.invoices.length">
              <td colspan="9" class="empty-state"><div class="icon"></div><h3>No invoices found</h3></td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="pagination" v-if="store.pagination.totalPages > 1">
        <button @click="loadInvoices(store.pagination.page - 1)" :disabled="store.pagination.page <= 1">Prev</button>
        <span class="text-muted">Page {{ store.pagination.page }} of {{ store.pagination.totalPages }}</span>
        <button @click="loadInvoices(store.pagination.page + 1)" :disabled="store.pagination.page >= store.pagination.totalPages">Next</button>
      </div>
    </template>

    <!-- Invoice Viewer Modal -->
    <Modal v-if="selectedInvoice" :title="`Invoice ${selectedInvoice.invoice_number}`" @close="selectedInvoice = null" maxWidth="800px">
      <div class="invoice-document">
        <div style="display:flex;justify-content:space-between;margin-bottom:30px">
          <div>
            <h2 style="color:#1a1a1a;font-size:1.5rem;margin-bottom:4px">INVOICE</h2>
            <p style="color:#6b7280;font-size:0.85rem">{{ selectedInvoice.invoice_number }}</p>
          </div>
          <div style="text-align:right">
            <h3 style="color:#4f46e5;font-size:1.2rem">SalesHub</h3>
            <p style="color:#6b7280;font-size:0.8rem">Sales Module System</p>
          </div>
        </div>

        <div style="display:flex;justify-content:space-between;margin-bottom:24px">
          <div>
            <p style="color:#6b7280;font-size:0.75rem;text-transform:uppercase;margin-bottom:4px">Bill To</p>
            <p style="font-weight:600">{{ selectedInvoice.customer_name }}</p>
          </div>
          <div style="text-align:right">
            <p style="color:#6b7280;font-size:0.8rem">Issue: {{ selectedInvoice.issue_date }}</p>
            <p style="color:#6b7280;font-size:0.8rem">Due: {{ selectedInvoice.due_date || 'N/A' }}</p>
          </div>
        </div>

        <table>
          <thead>
            <tr><th>Description</th><th>Qty</th><th>Unit Price</th><th style="text-align:right">Total</th></tr>
          </thead>
          <tbody>
            <tr v-for="item in selectedInvoice.items" :key="item.id">
              <td>{{ item.description }}</td>
              <td>{{ item.quantity }}</td>
              <td>Rs. {{ fmt(item.unit_price) }}</td>
              <td style="text-align:right;font-weight:600">Rs. {{ fmt(item.line_total) }}</td>
            </tr>
          </tbody>
        </table>

        <div style="margin-top:20px;text-align:right;border-top:2px solid #e5e7eb;padding-top:16px">
          <p style="color:#6b7280">Subtotal: Rs. {{ fmt(selectedInvoice.subtotal) }}</p>
          <p style="color:#6b7280">Discount: Rs. {{ fmt(selectedInvoice.discount) }}</p>
          <p style="font-size:1.2rem;font-weight:700;color:#4f46e5;margin-top:8px">Total: Rs. {{ fmt(selectedInvoice.total) }}</p>
        </div>
      </div>
      <template #footer>
        <button class="btn btn-ghost" @click="selectedInvoice = null">Close</button>
        <button class="btn btn-primary" @click="printInvoice">Print</button>
      </template>
    </Modal>

    <!-- Generate Invoice Modal -->
    <Modal v-if="showGenerateModal" title="Generate Invoice" @close="showGenerateModal = false" maxWidth="500px">
      <form @submit.prevent="submitGenerate">
        <div class="form-group">
          <label class="form-label">Customer</label>
          <select v-model="selectedCustomer" class="form-control" required @change="fetchCustomerSales">
            <option value="">Select Customer</option>
            <option v-for="c in customers" :key="c.id" :value="c.id">{{ c.name }}</option>
          </select>
        </div>
        <div class="form-group" v-if="selectedCustomer">
          <label class="form-label">Sale</label>
          <select v-model="selectedSale" class="form-control" required>
            <option value="">Select Sale</option>
            <option v-for="s in sales" :key="s.id" :value="s.id">
              {{ s.order_id }} - Rs. {{ fmt(s.total_amount) }} ({{ s.status }})
            </option>
          </select>
          <p v-if="sales.length === 0" class="text-muted" style="margin-top: 8px; font-size: 0.85rem;">No sales found for this customer.</p>
        </div>
      </form>
      <template #footer>
        <button class="btn btn-ghost" @click="showGenerateModal = false">Cancel</button>
        <button class="btn btn-primary" @click="submitGenerate" :disabled="!selectedSale || generating">
          {{ generating ? 'Generating...' : 'Generate' }}
        </button>
      </template>
    </Modal>
  </div>
</template>

<script>
import { useInvoicesStore } from '../stores/invoices'
import StatusBadge from '../components/common/StatusBadge.vue'
import LoadingSpinner from '../components/common/LoadingSpinner.vue'
import Modal from '../components/common/Modal.vue'

import api from '../services/api'

export default {
  name: 'InvoicesView',
  components: { StatusBadge, LoadingSpinner, Modal },
  data() { return { statusFilter: '', selectedInvoice: null, showGenerateModal: false, customers: [], sales: [], selectedCustomer: '', selectedSale: '', generating: false } },
  computed: { store() { return useInvoicesStore() } },
  methods: {
    async updateStatus(id, status) {
      if (!confirm(`Mark invoice as ${status}?`)) return
      try {
        await this.store.updateStatus(id, status)
        alert("Status updated")
      } catch (err) {
        alert("Failed to update status")
      }
    },
    async deleteInvoice(id) {
      if (!confirm("Delete this invoice forever?")) return
      try {
        await this.store.deleteInvoice(id)
        alert("Invoice deleted")
      } catch (err) {
        alert("Failed to delete invoice")
      }
    },
    fmt(n) { return Number(n || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) },
    loadInvoices(page = 1) {
      this.store.fetchInvoices(page, { status: this.statusFilter })
    },
    async viewInvoice(id) {
      this.selectedInvoice = await this.store.fetchInvoice(id)
    },
    printInvoice() { window.print() },
    async openGenerateModal() {
      this.showGenerateModal = true
      this.selectedCustomer = ''
      this.selectedSale = ''
      this.sales = []
      try {
        const { data } = await api.get('/customers')
        this.customers = data || []
      } catch (e) {
        console.error('Failed to load customers', e)
      }
    },
    async fetchCustomerSales() {
      this.selectedSale = ''
      if (!this.selectedCustomer) {
        this.sales = []
        return
      }
      try {
        const { data } = await api.get('/sales', { params: { customer_id: this.selectedCustomer, limit: 100 } })
        this.sales = (data.data || []).filter(s => s.status !== 'cancelled')
      } catch (e) {
        console.error('Failed to fetch sales', e)
      }
    },
    async submitGenerate() {
      if (!this.selectedSale) return
      this.generating = true
      try {
        await this.store.generateInvoice(this.selectedSale)
        alert('Invoice generated successfully!')
        this.showGenerateModal = false
        this.loadInvoices(1)
      } catch (err) {
        alert(err.response?.data?.error || 'Failed to generate invoice')
      } finally {
        this.generating = false
      }
    }
  },
  mounted() { this.loadInvoices() }
}
</script>
