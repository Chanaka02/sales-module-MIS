package services

import (
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"

	"sales-module/internal/models"
)

type InvoiceService struct {
	db *sql.DB
}

func NewInvoiceService(db *sql.DB) *InvoiceService {
	return &InvoiceService{db: db}
}

func (s *InvoiceService) ListInvoices(filter models.InvoiceFilter) (*models.PaginatedResponse, error) {
	if filter.Page < 1 { filter.Page = 1 }
	if filter.Limit < 1 || filter.Limit > 100 { filter.Limit = 20 }

	conditions := []string{}
	args := []interface{}{}
	argIdx := 1

	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("i.status = $%d", argIdx)); args = append(args, filter.Status); argIdx++
	}
	if filter.CustomerID > 0 {
		conditions = append(conditions, fmt.Sprintf("i.customer_id = $%d", argIdx)); args = append(args, filter.CustomerID); argIdx++
	}
	if filter.DateFrom != "" {
		conditions = append(conditions, fmt.Sprintf("i.issue_date >= $%d", argIdx)); args = append(args, filter.DateFrom); argIdx++
	}
	if filter.DateTo != "" {
		conditions = append(conditions, fmt.Sprintf("i.issue_date <= $%d", argIdx)); args = append(args, filter.DateTo); argIdx++
	}

	where := ""
	if len(conditions) > 0 { where = "WHERE " + strings.Join(conditions, " AND ") }

	var total int
	s.db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM invoices i %s", where), args...).Scan(&total)

	offset := (filter.Page - 1) * filter.Limit
	query := fmt.Sprintf(`SELECT i.id, i.sale_id, i.invoice_number, i.customer_id, COALESCE(c.name,''),
		i.issue_date, COALESCE(i.due_date::text,''), i.subtotal, i.discount, i.tax, i.total, i.status, i.created_at, i.updated_at
		FROM invoices i LEFT JOIN customers c ON i.customer_id=c.id %s ORDER BY i.created_at DESC LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args = append(args, filter.Limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil { return nil, err }
	defer rows.Close()

	invoices := []models.Invoice{}
	for rows.Next() {
		var inv models.Invoice
		rows.Scan(&inv.ID, &inv.SaleID, &inv.InvoiceNumber, &inv.CustomerID, &inv.CustomerName,
			&inv.IssueDate, &inv.DueDate, &inv.Subtotal, &inv.Discount, &inv.Tax, &inv.Total, &inv.Status, &inv.CreatedAt, &inv.UpdatedAt)
		invoices = append(invoices, inv)
	}

	return &models.PaginatedResponse{Data: invoices, Total: total, Page: filter.Page, Limit: filter.Limit, TotalPages: int(math.Ceil(float64(total) / float64(filter.Limit)))}, nil
}

func (s *InvoiceService) GetInvoice(id int) (*models.Invoice, error) {
	inv := &models.Invoice{}
	err := s.db.QueryRow(`SELECT i.id, i.sale_id, i.invoice_number, i.customer_id, COALESCE(c.name,''),
		i.issue_date, COALESCE(i.due_date::text,''), i.subtotal, i.discount, i.tax, i.total, i.status, i.created_at, i.updated_at
		FROM invoices i LEFT JOIN customers c ON i.customer_id=c.id WHERE i.id=$1`, id).Scan(
		&inv.ID, &inv.SaleID, &inv.InvoiceNumber, &inv.CustomerID, &inv.CustomerName,
		&inv.IssueDate, &inv.DueDate, &inv.Subtotal, &inv.Discount, &inv.Tax, &inv.Total, &inv.Status, &inv.CreatedAt, &inv.UpdatedAt)
	if err == sql.ErrNoRows { return nil, fmt.Errorf("invoice not found") }
	if err != nil { return nil, err }

	rows, _ := s.db.Query(`SELECT ii.id, ii.invoice_id, ii.product_id, ii.description, ii.quantity, ii.unit_price, ii.discount, ii.line_total
		FROM invoice_items ii WHERE ii.invoice_id=$1 ORDER BY ii.id`, id)
	defer rows.Close()
	for rows.Next() {
		var item models.InvoiceItem
		rows.Scan(&item.ID, &item.InvoiceID, &item.ProductID, &item.Description, &item.Quantity, &item.UnitPrice, &item.Discount, &item.LineTotal)
		inv.Items = append(inv.Items, item)
	}
	return inv, nil
}

func (s *InvoiceService) GenerateInvoice(saleID int) (*models.Invoice, error) {
	var cnt int
	s.db.QueryRow("SELECT COUNT(*) FROM invoices WHERE sale_id=$1", saleID).Scan(&cnt)
	if cnt > 0 { return nil, fmt.Errorf("invoice already exists for this sale") }

	var sale models.Sale
	err := s.db.QueryRow(`SELECT id, customer_id, subtotal, discount_amount, tax_amount, total_amount, status FROM sales WHERE id=$1`, saleID).Scan(
		&sale.ID, &sale.CustomerID, &sale.Subtotal, &sale.DiscountAmount, &sale.TaxAmount, &sale.TotalAmount, &sale.Status)
	if err == sql.ErrNoRows { return nil, fmt.Errorf("sale not found") }
	if err != nil { return nil, err }
	if sale.Status != models.SaleStatusCompleted { return nil, fmt.Errorf("can only generate invoices for completed sales") }

	invoiceNumber := fmt.Sprintf("INV-%s-%04d", time.Now().Format("20060102"), cnt+1)
	dueDate := time.Now().AddDate(0, 0, 30)

	tx, _ := s.db.Begin()
	defer tx.Rollback()

	var invoiceID int
	tx.QueryRow(`INSERT INTO invoices (sale_id, invoice_number, customer_id, issue_date, due_date, subtotal, discount, tax, total, status)
		VALUES ($1,$2,$3,CURRENT_DATE,$4,$5,$6,$7,$8,'draft') RETURNING id`,
		saleID, invoiceNumber, sale.CustomerID, dueDate, sale.Subtotal, sale.DiscountAmount, sale.TaxAmount, sale.TotalAmount).Scan(&invoiceID)

	rows, _ := s.db.Query(`SELECT si.product_id, COALESCE(p.name,''), si.quantity, si.unit_price, si.line_total
		FROM sale_items si LEFT JOIN products p ON si.product_id=p.id WHERE si.sale_id=$1`, saleID)
	defer rows.Close()
	for rows.Next() {
		var pid *int; var desc string; var qty int; var up, lt float64
		rows.Scan(&pid, &desc, &qty, &up, &lt)
		tx.Exec(`INSERT INTO invoice_items (invoice_id,product_id,description,quantity,unit_price,discount,line_total) VALUES($1,$2,$3,$4,$5,0.00,$6)`,
			invoiceID, pid, desc, qty, up, lt)
	}
	tx.Commit()
	return s.GetInvoice(invoiceID)
}

func (s *InvoiceService) Delete(id int) error {
        _, err := s.db.Exec(`DELETE FROM invoices WHERE id=$1`, id)
        return err
}

func (s *InvoiceService) UpdateStatus(id int, status string) error {
        _, err := s.db.Exec(`UPDATE invoices SET status=$1, updated_at=CURRENT_TIMESTAMP WHERE id=$2`, status, id)
        return err
}
