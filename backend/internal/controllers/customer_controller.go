package controllers

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"sales-module/internal/models"
)

type CustomerController struct {
	db *sql.DB
}

func NewCustomerController(db *sql.DB) *CustomerController {
	return &CustomerController{db: db}
}

func (cc *CustomerController) List(c *gin.Context) {
	rows, err := cc.db.Query(`SELECT c.id, c.name, COALESCE(c.email,''), COALESCE(c.phone,''), COALESCE(c.address,''), COALESCE(c.company,''),
		COALESCE(sc.cnt,0), COALESCE(sc.total,0), c.created_at, c.updated_at
		FROM customers c LEFT JOIN (SELECT customer_id, COUNT(*) as cnt, SUM(total_amount) as total FROM sales WHERE status='completed' GROUP BY customer_id) sc ON c.id=sc.customer_id
		ORDER BY c.name`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	customers := []models.Customer{}
	for rows.Next() {
		var cust models.Customer
		rows.Scan(&cust.ID, &cust.Name, &cust.Email, &cust.Phone, &cust.Address, &cust.Company, &cust.TotalSales, &cust.TotalSpent, &cust.CreatedAt, &cust.UpdatedAt)
		customers = append(customers, cust)
	}
	c.JSON(http.StatusOK, customers)
}

func (cc *CustomerController) GetHistory(c *gin.Context) {
	custID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 { page = 1 }
	if limit < 1 || limit > 100 { limit = 20 }

	conditions := []string{"s.customer_id = $1"}
	args := []interface{}{custID}
	argIdx := 2

	if status := c.Query("status"); status != "" {
		conditions = append(conditions, fmt.Sprintf("s.status = $%d", argIdx)); args = append(args, status); argIdx++
	}
	if df := c.Query("date_from"); df != "" {
		conditions = append(conditions, fmt.Sprintf("s.created_at >= $%d", argIdx)); args = append(args, df); argIdx++
	}
	if dt := c.Query("date_to"); dt != "" {
		conditions = append(conditions, fmt.Sprintf("s.created_at <= $%d", argIdx)); args = append(args, dt+"T23:59:59"); argIdx++
	}
	if pid := c.Query("product_id"); pid != "" {
		conditions = append(conditions, fmt.Sprintf("EXISTS(SELECT 1 FROM sale_items si WHERE si.sale_id=s.id AND si.product_id=$%d)", argIdx))
		args = append(args, pid); argIdx++
	}

	where := "WHERE " + strings.Join(conditions, " AND ")
	var total int
	cc.db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM sales s %s", where), args...).Scan(&total)

	offset := (page - 1) * limit
	query := fmt.Sprintf(`SELECT s.id, s.customer_id, COALESCE(cu.name,''), s.order_id, s.status, s.subtotal, s.discount_percent, s.discount_amount, s.tax_amount, s.total_amount, COALESCE(s.notes,''), s.created_at, s.updated_at
		FROM sales s LEFT JOIN customers cu ON s.customer_id=cu.id %s ORDER BY s.created_at DESC LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := cc.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	sales := []models.Sale{}
	for rows.Next() {
		var sale models.Sale
		rows.Scan(&sale.ID, &sale.CustomerID, &sale.CustomerName, &sale.OrderID, &sale.Status, &sale.Subtotal, &sale.DiscountPercent, &sale.DiscountAmount, &sale.TaxAmount, &sale.TotalAmount, &sale.Notes, &sale.CreatedAt, &sale.UpdatedAt)
		sales = append(sales, sale)
	}

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Data: sales, Total: total, Page: page, Limit: limit,
		TotalPages: int(math.Ceil(float64(total) / float64(limit))),
	})
}

func (cc *CustomerController) Create(c *gin.Context) {
        var cust models.Customer
        if err := c.ShouldBindJSON(&cust); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
        err := cc.db.QueryRow(`INSERT INTO customers (name, email, phone, address, company) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at`, cust.Name, cust.Email, cust.Phone, cust.Address, cust.Company).Scan(&cust.ID, &cust.CreatedAt, &cust.UpdatedAt)
        if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
        c.JSON(http.StatusCreated, cust)
}

func (cc *CustomerController) Update(c *gin.Context) {
        id := c.Param("id")
        var cust models.Customer
        if err := c.ShouldBindJSON(&cust); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
        _, err := cc.db.Exec(`UPDATE customers SET name=$1, email=$2, phone=$3, address=$4, company=$5, updated_at=CURRENT_TIMESTAMP WHERE id=$6`, cust.Name, cust.Email, cust.Phone, cust.Address, cust.Company, id)
        if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
        c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (cc *CustomerController) Delete(c *gin.Context) {
        if _, err := cc.db.Exec(`DELETE FROM customers WHERE id=$1`, c.Param("id")); err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
        c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
