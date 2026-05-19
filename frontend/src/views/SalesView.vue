<template>
  <div class="sales-page">
    <div class="page-header">
      <div>
        <h1>Sales Transactions</h1>
        <p>Manage all your sales orders</p>
      </div>
      <button class="btn btn-primary" @click="openCreateModal" id="create-sale-btn">New Sale</button>
    </div>

    <!-- Filters -->
    <div class="filters-bar">
      <select v-model="filters.status" @change="applyFilters" class="form-control" id="filter-status">
        <option value="">All Status</option>
        <option value="pending">Pending</option>
        <option value="completed">Completed</option>
        <option value="cancelled">Cancelled</option>
      </select>
      <input v-model="filters.date_from" type="date" class="form-control" @change="applyFilters" placeholder="From" />
      <input v-model="filters.date_to" type="date" class="form-control" @change="applyFilters" placeholder="To" />
      <button class="btn btn-ghost btn-sm" @click="clearFilters">Clear</button>
    </div>

    <LoadingSpinner v-if="store.loading" />

    <template v-else>
      <div class="card">
        <table class="data-table">
          <thead>
            <tr>
              <th>Order ID</th>
              <th>Customer</th>
              <th>Items</th>
              <th>Subtotal</th>
              <th>Discount</th>
              <th>Total</th>
              <th>Status</th>
              <th>Date</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="sale in store.sales" :key="sale.id" class="sale-row">
              <td class="text-primary font-medium">{{ sale.order_id }}</td>
              <td>{{ sale.customer_name || 'N/A' }}</td>
              <td class="text-muted">-</td>
              <td>Rs. {{ fmt(sale.subtotal) }}</td>
              <td class="text-warning" v-if="sale.discount_percent > 0">{{ sale.discount_percent }}%</td>
              <td v-else class="text-muted">-</td>
              <td class="text-primary font-semibold">Rs. {{ fmt(sale.total_amount) }}</td>
              <td><StatusBadge :status="sale.status" /></td>
              <td class="text-muted">{{ formatDate(sale.created_at) }}</td>
              <td>
                <div class="action-btns">
                  <button class="btn btn-ghost btn-sm" @click="viewSale(sale.id)" title="View">View</button>
                  <button v-if="sale.status === 'pending'" class="btn btn-success btn-sm" @click="completeSale(sale.id)" title="Complete">Complete</button>
                  <button v-if="sale.status === 'pending'" class="btn btn-danger btn-sm" @click="cancelSale(sale.id)" title="Cancel">Cancel</button>
                </div>
              </td>
            </tr>
            <tr v-if="!store.sales.length">
              <td colspan="9" class="empty-state">
                <div class="icon"></div>
                <h3>No sales found</h3>
                <p>Create your first sale to get started</p>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination -->
      <div class="pagination" v-if="store.pagination.totalPages > 1">
        <button @click="goPage(store.pagination.page - 1)" :disabled="store.pagination.page <= 1">Prev</button>
        <button v-for="p in visiblePages" :key="p" @click="goPage(p)" :class="{ active: p === store.pagination.page }">{{ p }}</button>
        <button @click="goPage(store.pagination.page + 1)" :disabled="store.pagination.page >= store.pagination.totalPages">Next</button>
      </div>
    </template>

    <!-- View Sale Modal -->
    <Modal v-if="selectedSale" :title="`Sale ${selectedSale.order_id}`" @close="selectedSale = null" maxWidth="700px">
      <div class="sale-detail">
        <div class="detail-grid">
          <div><span class="detail-label">Customer</span><span>{{ selectedSale.customer_name || 'N/A' }}</span></div>
          <div><span class="detail-label">Status</span><StatusBadge :status="selectedSale.status" /></div>
          <div><span class="detail-label">Subtotal</span><span>Rs. {{ fmt(selectedSale.subtotal) }}</span></div>
          <div><span class="detail-label">Discount</span><span>{{ selectedSale.discount_percent }}% (Rs. {{ fmt(selectedSale.discount_amount) }})</span></div>
          <div><span class="detail-label">Total</span><span style="font-weight:700;color:var(--color-primary-light)">Rs. {{ fmt(selectedSale.total_amount) }}</span></div>
          <div><span class="detail-label">Notes</span><span>{{ selectedSale.notes || '-' }}</span></div>
        </div>
        <h4 style="margin: 20px 0 12px; color: var(--text-primary)">Items</h4>
        <table class="data-table">
          <thead><tr><th>Product</th><th>SKU</th><th>Qty</th><th>Price</th><th>Total</th></tr></thead>
          <tbody>
            <tr v-for="item in selectedSale.items" :key="item.id">
              <td>{{ item.product_name }}</td>
              <td class="text-muted">{{ item.product_sku }}</td>
              <td>{{ item.quantity }}</td>
              <td>Rs. {{ fmt(item.unit_price) }}</td>
              <td style="font-weight:600">Rs. {{ fmt(item.line_total) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </Modal>

    <!-- Create Sale Modal -->
    <Modal v-if="showCreateModal" title="Create New Sale" @close="showCreateModal = false" maxWidth="700px">
      <form @submit.prevent="createSale">
        <div class="form-group">
          <label class="form-label">Customer</label>
          <select v-model="newSale.customer_id" class="form-control" required>
            <option value="">Select Customer</option>
            <option v-for="c in customers" :key="c.id" :value="c.id">{{ c.name }}</option>
          </select>
        </div>
        <div class="form-group">
          <label class="form-label">Discount (%)</label>
          <input v-model.number="newSale.discount_percent" type="number" min="0" max="50" step="0.5" class="form-control" />
        </div>
        <div class="form-group">
          <label class="form-label">Notes</label>
          <input v-model="newSale.notes" type="text" class="form-control" placeholder="Optional notes" />
        </div>

        <h4 style="margin: 16px 0 12px; color: var(--text-primary)">Items</h4>
        <div v-for="(item, i) in newSale.items" :key="i" class="item-row">
          <select v-model="item.product_id" class="form-control" required @change="setPrice(i)">
            <option value="">Product</option>
            <option v-for="p in products" :key="p.id" :value="p.id">{{ p.name }} (Rs. {{ fmt(p.price) }})</option>
          </select>
          <input v-model.number="item.quantity" type="number" min="1" class="form-control" placeholder="Qty" style="width:80px" required />
          <input v-model.number="item.unit_price" type="number" min="0" step="0.01" class="form-control" placeholder="Price" style="width:130px" required />
          <button type="button" class="btn btn-danger btn-sm" @click="removeItem(i)" v-if="newSale.items.length > 1">X</button>
        </div>
        <button type="button" class="btn btn-ghost btn-sm mt-1" @click="addItem">Add Item</button>

      </form>

      <template #footer>
        <button class="btn btn-ghost" @click="showCreateModal = false">Cancel</button>
        <button class="btn btn-primary" @click="createSale" :disabled="creating">{{ creating ? 'Creating...' : 'Create Sale' }}</button>
      </template>
    </Modal>
  </div>
</template>

<script>
import { useSalesStore } from '../stores/sales'
import StatusBadge from '../components/common/StatusBadge.vue'
import LoadingSpinner from '../components/common/LoadingSpinner.vue'
import Modal from '../components/common/Modal.vue'
import api from '../services/api'

export default {
  name: 'SalesView',
  components: { StatusBadge, LoadingSpinner, Modal },
  data() {
    return {
      filters: { status: '', date_from: '', date_to: '' },
      selectedSale: null,
      showCreateModal: false,
      creating: false,
      customers: [],
      products: [],
      newSale: { customer_id: '', discount_percent: 0, notes: '', items: [{ product_id: '', quantity: 1, unit_price: 0 }] }
    }
  },
  computed: {
    store() { return useSalesStore() },
    visiblePages() {
      const p = this.store.pagination
      const pages = []
      const start = Math.max(1, p.page - 2)
      const end = Math.min(p.totalPages, p.page + 2)
      for (let i = start; i <= end; i++) pages.push(i)
      return pages
    }
  },
  methods: {
    fmt(n) { return Number(n || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) },
    formatDate(d) { return new Date(d).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' }) },
    applyFilters() { this.store.filters = { ...this.filters }; this.store.fetchSales(1) },
    clearFilters() { this.filters = { status: '', date_from: '', date_to: '' }; this.applyFilters() },
    goPage(p) { this.store.fetchSales(p) },
    async viewSale(id) {
      const sale = await this.store.fetchSale(id)
      this.selectedSale = sale
    },
    async completeSale(id) {
      if (confirm('Mark this sale as completed?')) {
        await this.store.updateStatus(id, 'completed')
        this.store.fetchSales(this.store.pagination.page)
      }
    },
    async cancelSale(id) {
      if (confirm('Cancel this sale?')) {
        await this.store.updateStatus(id, 'cancelled')
        this.store.fetchSales(this.store.pagination.page)
      }
    },
    addItem() { this.newSale.items.push({ product_id: '', quantity: 1, unit_price: 0 }) },
    removeItem(i) { this.newSale.items.splice(i, 1) },
    setPrice(i) {
      const p = this.products.find(p => p.id === this.newSale.items[i].product_id)
      if (p) this.newSale.items[i].unit_price = p.price
    },
    async createSale() {
      if (!this.newSale.customer_id) return alert('Please select a customer')
      if (this.newSale.items.some(i => !i.product_id || i.quantity < 1)) return alert('Please select products and quantities')
      
      this.creating = true
      try {
        const payload = { 
          customer_id: Number(this.newSale.customer_id), 
          discount_percent: Number(this.newSale.discount_percent || 0),
          notes: this.newSale.notes,
          items: this.newSale.items.map(i => ({ 
            product_id: Number(i.product_id), 
            quantity: Number(i.quantity), 
            unit_price: Number(i.unit_price) 
          })) 
        }
        await this.store.createSale(payload)
        this.showCreateModal = false
        this.newSale = { customer_id: '', discount_percent: 0, notes: '', items: [{ product_id: '', quantity: 1, unit_price: 0 }] }
        this.store.fetchSales(1)
      } catch (err) {
        alert(err.response?.data?.error || 'Failed to create sale')
      } finally { this.creating = false }
    },
    async openCreateModal() {
      this.showCreateModal = true
      await this.loadMasterData()
    },
    async loadMasterData() {
      try {
        const [custRes, prodRes] = await Promise.all([
          api.get('/customers'),
          api.get('/products')
        ])
        this.customers = custRes.data || []
        this.products = prodRes.data || []
      } catch (e) { 
        console.error('Failed to load master data', e)
      }
    }
  },
  async mounted() {
    this.store.fetchSales(1)
    this.loadMasterData()
  }
}
</script>

<style scoped>
.action-btns { display: flex; gap: 6px; }

.sale-detail .detail-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.detail-label {
  display: block;
  font-size: 0.75rem;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 4px;
}

.item-row {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
  align-items: center;
}

.item-row .form-control { flex: 1; }
</style>
