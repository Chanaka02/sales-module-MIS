<template>
  <div class="customers-page">
    <div class="page-header" style="display:flex; justify-content:space-between; align-items:center;">
      <div>
        <h1>Customers</h1>
        <p>Customer directory and sales history</p>
      </div>
      <button class="btn btn-primary" @click="openCustomerModal()"><i class='bx bx-plus'></i> Add Customer</button>
    </div>

    <div class="filters-bar" style="margin-bottom:20px;">
      <input v-model="search" type="text" class="form-control" placeholder="Search customers..." style="max-width:300px" />
    </div>

    <LoadingSpinner v-if="loading" />

    <template v-else>
      <div class="card">
        <table class="data-table">
          <thead>
            <tr>
              <th>Name</th>
              <th>Email</th>
              <th>Phone</th>
              <th>Company</th>
              <th>Total Sales</th>
              <th>Total Spent</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="c in filteredCustomers" :key="c.id">
              <td class="text-primary font-medium">{{ c.name }}</td>
              <td class="text-muted">{{ c.email || '-' }}</td>
              <td class="text-muted">{{ c.phone || '-' }}</td>
              <td>{{ c.company || '-' }}</td>
              <td>{{ c.total_sales }}</td>
              <td class="text-success font-semibold">Rs. {{ fmt(c.total_spent) }}</td>
              <td style="display:flex; gap:8px;">
                <button class="btn btn-ghost btn-sm" @click="viewHistory(c)" title="View History"><i class='bx bx-history'></i></button>
                <button class="btn btn-ghost btn-sm" @click="openCustomerModal(c)" title="Edit"><i class='bx bx-edit'></i></button>
                <button class="btn btn-danger btn-sm" @click="deleteCustomer(c.id)" title="Delete" style="background:#fee2e2; color:#ef4444; border:none; padding:4px 8px; border-radius:4px; cursor:pointer;"><i class='bx bx-trash'></i></button>
              </td>
            </tr>
            <tr v-if="!filteredCustomers.length">
                <td colspan="7" class="text-center text-muted" style="padding:20px;">No customers found.</td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>

    <!-- Customer History Modal -->
    <Modal v-if="selectedCustomer" :title="`${selectedCustomer.name} — Sales History`" @close="selectedCustomer = null" maxWidth="800px">
      <div class="filters-bar mb-2">
        <select v-model="historyFilter.status" @change="loadHistory" class="form-control">
          <option value="">All Status</option>
          <option value="pending">Pending</option>
          <option value="completed">Completed</option>
          <option value="cancelled">Cancelled</option>
        </select>
        <input v-model="historyFilter.date_from" type="date" class="form-control" @change="loadHistory" />
        <input v-model="historyFilter.date_to" type="date" class="form-control" @change="loadHistory" />
      </div>

      <LoadingSpinner v-if="historyLoading" message="Loading history..." />
      <table v-else class="data-table">
        <thead>
          <tr><th>Order</th><th>Status</th><th>Total</th><th>Date</th></tr>
        </thead>
        <tbody>
          <tr v-for="s in history" :key="s.id">
            <td class="text-primary font-medium">{{ s.order_id }}</td>
            <td><StatusBadge :status="s.status" /></td>
            <td style="font-weight:600">Rs. {{ fmt(s.total_amount) }}</td>
            <td class="text-muted">{{ new Date(s.created_at).toLocaleDateString() }}</td>
          </tr>
          <tr v-if="!history.length">
            <td colspan="4" class="text-center text-muted" style="padding:30px">No sales history found</td>
          </tr>
        </tbody>
      </table>
    </Modal>

    <!-- Add/Edit Customer Modal -->
    <Modal v-if="showCustomerModal" :title="editingCustomer ? 'Edit Customer' : 'Add New Customer'" @close="closeCustomerModal" maxWidth="500px">
      <form @submit.prevent="saveCustomer">
        <div class="form-group" style="margin-bottom:15px;">
          <label class="form-label" style="display:block;margin-bottom:5px;font-weight:600;font-size:0.85rem">Full Name *</label>
          <input v-model="customerForm.name" type="text" class="form-control" required placeholder="John Doe" style="width:100%;padding:8px;border:1px solid var(--border-color);border-radius:4px" />
        </div>
        <div class="form-group" style="margin-bottom:15px;">
          <label class="form-label" style="display:block;margin-bottom:5px;font-weight:600;font-size:0.85rem">Email</label>
          <input v-model="customerForm.email" type="email" class="form-control" placeholder="john@example.com" style="width:100%;padding:8px;border:1px solid var(--border-color);border-radius:4px" />
        </div>
        <div class="form-group" style="margin-bottom:15px;">
          <label class="form-label" style="display:block;margin-bottom:5px;font-weight:600;font-size:0.85rem">Phone</label>
          <input v-model="customerForm.phone" type="text" class="form-control" placeholder="+1 555-0199" style="width:100%;padding:8px;border:1px solid var(--border-color);border-radius:4px" />
        </div>
        <div class="form-group" style="margin-bottom:15px;">
          <label class="form-label" style="display:block;margin-bottom:5px;font-weight:600;font-size:0.85rem">Company</label>
          <input v-model="customerForm.company" type="text" class="form-control" placeholder="Acme Corp" style="width:100%;padding:8px;border:1px solid var(--border-color);border-radius:4px" />
        </div>
        <div class="form-group" style="margin-bottom:20px;">
          <label class="form-label" style="display:block;margin-bottom:5px;font-weight:600;font-size:0.85rem">Address</label>
          <textarea v-model="customerForm.address" class="form-control" rows="3" placeholder="123 Main St..." style="width:100%;padding:8px;border:1px solid var(--border-color);border-radius:4px"></textarea>
        </div>
        <div style="display:flex;justify-content:flex-end;gap:10px;padding-top:15px;border-top:1px solid var(--border-color)">
          <button type="button" class="btn" @click="closeCustomerModal" style="padding:8px 16px;background:var(--bg-secondary);border:1px solid var(--border-color);border-radius:4px">Cancel</button>
          <button type="submit" class="btn btn-primary" :disabled="saving" style="padding:8px 16px;background:var(--color-primary);color:white;border:none;border-radius:4px">
            {{ saving ? (editingCustomer ? 'Updating...' : 'Adding...') : (editingCustomer ? 'Update Customer' : 'Add Customer') }}
          </button>
        </div>
      </form>
    </Modal>
  </div>
</template>

<script>
import api from '../services/api'
import StatusBadge from '../components/common/StatusBadge.vue'
import LoadingSpinner from '../components/common/LoadingSpinner.vue'
import Modal from '../components/common/Modal.vue'

export default {
  name: 'CustomersView',
  components: { StatusBadge, LoadingSpinner, Modal },
  data() {
    return {
      customers: [], loading: true, search: '',
      selectedCustomer: null, history: [], historyLoading: false,
      historyFilter: { status: '', date_from: '', date_to: '' },
      
      // Add/Edit state
      showCustomerModal: false,
      editingCustomer: null,
      saving: false,
      customerForm: { name: '', email: '', phone: '', company: '', address: '' }
    }
  },
  computed: {
    filteredCustomers() {
      if (!this.search) return this.customers
      const q = this.search.toLowerCase()
      return this.customers.filter(c => c.name.toLowerCase().includes(q) || (c.email || '').toLowerCase().includes(q) || (c.company || '').toLowerCase().includes(q))
    }
  },
  methods: {
    fmt(n) { return Number(n || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) },
    async loadCustomers() {
      this.loading = true
      try {
        const { data } = await api.get('/customers')
        this.customers = data || []
      } finally { this.loading = false }
    },
    async viewHistory(customer) {
      this.selectedCustomer = customer
      this.historyFilter = { status: '', date_from: '', date_to: '' }
      this.loadHistory()
    },
    async loadHistory() {
      this.historyLoading = true
      try {
        const params = { ...this.historyFilter, limit: 50 }
        Object.keys(params).forEach(k => { if (!params[k]) delete params[k] })
        const { data } = await api.get(`/customers/${this.selectedCustomer.id}/history`, { params })
        this.history = data.data || []
      } finally { this.historyLoading = false }
    },
    openCustomerModal(customer = null) {
      if (customer) {
        this.editingCustomer = customer
        this.customerForm = { ...customer }
      } else {
        this.editingCustomer = null
        this.customerForm = { name: '', email: '', phone: '', company: '', address: '' }
      }
      this.showCustomerModal = true
    },
    closeCustomerModal() {
      this.showCustomerModal = false
      this.editingCustomer = null
      this.customerForm = { name: '', email: '', phone: '', company: '', address: '' }
    },
    async saveCustomer() {
      this.saving = true
      try {
        if (this.editingCustomer) {
          await api.put(`/customers/${this.editingCustomer.id}`, this.customerForm)
          alert('Customer updated successfully')
        } else {
          await api.post('/customers', this.customerForm)
          alert('Customer added successfully')
        }
        this.closeCustomerModal()
        this.loadCustomers()
      } catch (err) {
        alert(err.response?.data?.error || 'Error saving customer')
      } finally {
        this.saving = false
      }
    },
    async deleteCustomer(id) {
      if (!confirm('Are you sure you want to delete this customer? This cannot be undone.')) return
      try {
        await api.delete(`/customers/${id}`)
        alert('Customer deleted')
        this.loadCustomers()
      } catch(err) {
        alert(err.response?.data?.error || 'Error deleting customer')
      }
    }
  },
  mounted() {
    this.loadCustomers()
  }
}
</script>
