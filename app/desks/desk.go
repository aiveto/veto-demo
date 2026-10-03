// Package desks is the sample services. Orders, customers, and billing speak the contracts in app/contracts.
package desks

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	OrdersAddr    = "127.0.0.1:18410"
	CustomersAddr = "127.0.0.1:18411"
	InvoicesAddr  = "127.0.0.1:18412"
	Token         = "demo-token"
)

type (
	Address struct {
		Name       string `json:"name"`
		Line1      string `json:"line1"`
		City       string `json:"city"`
		Region     string `json:"region"`
		PostalCode string `json:"postalCode"`
		Country    string `json:"country"`
	}

	Line struct {
		SKU        string `json:"sku"`
		Name       string `json:"name"`
		Quantity   int    `json:"quantity"`
		UnitAmount int    `json:"unitAmount"`
	}

	Order struct {
		ID          string  `json:"id"`
		Status      string  `json:"status"`
		Currency    string  `json:"currency"`
		CustomerID  string  `json:"customerId"`
		InvoiceID   string  `json:"invoiceId"`
		WarehouseID string  `json:"warehouseId"`
		PlacedAt    string  `json:"placedAt"`
		TotalAmount int     `json:"totalAmount"`
		Email       string  `json:"email"`
		Lines       []Line  `json:"lines"`
		ShipTo      Address `json:"shipTo"`
	}

	Customer struct {
		ID      string  `json:"id"`
		Name    string  `json:"name"`
		Email   string  `json:"email"`
		Phone   string  `json:"phone,omitempty"`
		Since   string  `json:"since"`
		Address Address `json:"address"`
	}

	Invoice struct {
		ID          string `json:"id"`
		OrderID     string `json:"orderId"`
		CustomerID  string `json:"customerId"`
		Status      string `json:"status"`
		Currency    string `json:"currency"`
		TotalAmount int    `json:"totalAmount"`
		IssuedAt    string `json:"issuedAt"`
	}

	Call struct {
		Service string `json:"service"`
		Method  string `json:"method"`
		Path    string `json:"path"`
		Status  int    `json:"status"`
	}

	Desk struct {
		mu        sync.Mutex
		orders    map[string]*Order
		customers map[string]*Customer
		invoices  map[string]*Invoice
		calls     []Call
	}
)

func New() *Desk {
	mara := Address{
		Name: "Mara Ellison", Line1: "418 Clement Street", City: "San Francisco",
		Region: "CA", PostalCode: "94118", Country: "US",
	}
	jonas := Address{
		Name: "Jonas Adler", Line1: "90 South 2nd Street", City: "Brooklyn",
		Region: "NY", PostalCode: "11249", Country: "US",
	}
	priya := Address{
		Name: "Priya Raman", Line1: "16th Avenue", City: "Seattle",
		Region: "WA", PostalCode: "98122", Country: "US",
	}
	return &Desk{
		orders: map[string]*Order{
			"10482": {
				ID: "10482", Status: "fulfilled", Currency: "USD",
				CustomerID: "cus_mara", InvoiceID: "inv_2291", WarehouseID: "wh_sfo_1",
				PlacedAt: "2026-09-18T15:04:00Z", TotalAmount: 55600,
				Email: "mara.ellison@example.com",
				Lines: []Line{
					{SKU: "WL-COAT-UM", Name: "Undyed wool coat", Quantity: 1, UnitAmount: 42800},
					{SKU: "CD-TRAY-01", Name: "Cedar tray", Quantity: 2, UnitAmount: 6400},
				},
				ShipTo: mara,
			},
			"10490": {
				ID: "10490", Status: "placed", Currency: "USD",
				CustomerID: "cus_jonas", InvoiceID: "inv_2304", WarehouseID: "wh_ewr_2",
				PlacedAt: "2026-09-28T11:20:00Z", TotalAmount: 18600,
				Email:  "jonas.adler@example.com",
				Lines:  []Line{{SKU: "LN-THRW-02", Name: "Linen throw", Quantity: 1, UnitAmount: 18600}},
				ShipTo: jonas,
			},
			"10502": {
				ID: "10502", Status: "fulfilled", Currency: "USD",
				CustomerID: "cus_priya", InvoiceID: "inv_2318", WarehouseID: "wh_sea_1",
				PlacedAt: "2026-09-30T18:41:00Z", TotalAmount: 9400,
				Email:  "priya.raman@example.com",
				Lines:  []Line{{SKU: "BR-MUG-04", Name: "Speckled mug", Quantity: 2, UnitAmount: 4700}},
				ShipTo: priya,
			},
		},
		customers: map[string]*Customer{
			"cus_mara":  {ID: "cus_mara", Name: "Mara Ellison", Email: "mara.ellison@example.com", Phone: "+1-415-555-0148", Since: "2022-04-03", Address: mara},
			"cus_jonas": {ID: "cus_jonas", Name: "Jonas Adler", Email: "jonas.adler@example.com", Phone: "+1-718-555-0194", Since: "2024-11-12", Address: jonas},
			"cus_priya": {ID: "cus_priya", Name: "Priya Raman", Email: "priya.raman@example.com", Since: "2025-06-01", Address: priya},
		},
		invoices: map[string]*Invoice{
			"inv_2291": {ID: "inv_2291", OrderID: "10482", CustomerID: "cus_mara", Status: "paid", Currency: "USD", TotalAmount: 55600, IssuedAt: "2026-09-18T15:04:11Z"},
			"inv_2304": {ID: "inv_2304", OrderID: "10490", CustomerID: "cus_jonas", Status: "open", Currency: "USD", TotalAmount: 18600, IssuedAt: "2026-09-28T11:20:08Z"},
			"inv_2318": {ID: "inv_2318", OrderID: "10502", CustomerID: "cus_priya", Status: "paid", Currency: "USD", TotalAmount: 9400, IssuedAt: "2026-09-30T18:41:06Z"},
		},
	}
}

func (d *Desk) Handler(service string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			writeJSON(w, http.StatusOK, map[string]string{"ok": "true", "service": service})
			return
		}
		if service == "orders" && r.URL.Path == "/_demo/log" && r.Method == http.MethodGet {
			d.mu.Lock()
			calls := append([]Call(nil), d.calls...)
			d.mu.Unlock()
			deletes := 0
			for _, c := range calls {
				if c.Method == http.MethodDelete {
					deletes++
				}
			}
			writeJSON(w, http.StatusOK, map[string]any{"deletes": deletes, "calls": calls})
			return
		}
		status := d.route(service, w, r)
		d.mu.Lock()
		d.calls = append(d.calls, Call{Service: service, Method: r.Method, Path: r.URL.Path, Status: status})
		d.mu.Unlock()
	})
}

func (d *Desk) route(service string, w http.ResponseWriter, r *http.Request) int {
	if r.Header.Get("Authorization") != "Bearer "+Token {
		return problem(w, http.StatusUnauthorized, "the desk wants its token")
	}
	switch service {
	case "orders":
		return d.ordersRoute(w, r)
	case "customers":
		return d.customersRoute(w, r)
	case "invoices":
		return d.invoicesRoute(w, r)
	default:
		return problem(w, http.StatusNotFound, "unknown desk")
	}
}

func (d *Desk) ordersRoute(w http.ResponseWriter, r *http.Request) int {
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")
	if path == "orders" && r.Method == http.MethodGet {
		want := r.URL.Query().Get("status")
		limit := pageSize(r.URL.Query().Get("limit"))
		d.mu.Lock()
		defer d.mu.Unlock()
		var data []map[string]any
		for _, id := range []string{"10502", "10490", "10482"} {
			o := d.orders[id]
			if o == nil {
				continue
			}
			if want != "" && o.Status != want {
				continue
			}
			data = append(data, map[string]any{"id": o.ID, "status": o.Status, "customerId": o.CustomerID, "totalAmount": o.TotalAmount})
			if len(data) == limit {
				break
			}
		}
		return writeJSON(w, http.StatusOK, map[string]any{"data": data, "nextCursor": nil})
	}
	if path == "orders" && r.Method == http.MethodPost {
		return writeJSON(w, http.StatusCreated, map[string]any{"id": "10510", "status": "placed"})
	}
	if len(parts) < 2 || parts[0] != "orders" {
		return problem(w, http.StatusNotFound, "no such order route")
	}
	id := parts[1]
	if len(parts) == 2 && r.Method == http.MethodGet {
		d.mu.Lock()
		o := d.orders[id]
		d.mu.Unlock()
		if o == nil {
			return problem(w, http.StatusNotFound, "order "+id+" is not on the desk")
		}
		return writeJSON(w, http.StatusOK, o)
	}
	if len(parts) == 2 && r.Method == http.MethodDelete {
		d.mu.Lock()
		_, ok := d.orders[id]
		delete(d.orders, id)
		d.mu.Unlock()
		if !ok {
			return problem(w, http.StatusNotFound, "order "+id+" is not on the desk")
		}
		w.WriteHeader(http.StatusNoContent)
		return http.StatusNoContent
	}
	if len(parts) == 2 && r.Method == http.MethodPatch {
		d.mu.Lock()
		o := d.orders[id]
		d.mu.Unlock()
		if o == nil {
			return problem(w, http.StatusNotFound, "order "+id+" is not on the desk")
		}
		return writeJSON(w, http.StatusOK, o)
	}
	if len(parts) == 3 && parts[2] == "cancel" && r.Method == http.MethodPost {
		d.mu.Lock()
		o := d.orders[id]
		if o != nil {
			o.Status = "cancelled"
		}
		d.mu.Unlock()
		if o == nil {
			return problem(w, http.StatusNotFound, "order "+id+" is not on the desk")
		}
		return writeJSON(w, http.StatusOK, o)
	}
	if len(parts) == 3 && parts[2] == "notes" && r.Method == http.MethodPost {
		if _, ok := d.lookupOrder(id); !ok {
			return problem(w, http.StatusNotFound, "order "+id+" is not on the desk")
		}
		return writeJSON(w, http.StatusCreated, map[string]string{"id": "note_" + id, "body": "noted", "at": time.Now().UTC().Format(time.RFC3339)})
	}
	if len(parts) == 3 && parts[2] == "events" && r.Method == http.MethodGet {
		if _, ok := d.lookupOrder(id); !ok {
			return problem(w, http.StatusNotFound, "order "+id+" is not on the desk")
		}
		return writeJSON(w, http.StatusOK, map[string]any{"data": []map[string]string{{"at": "2026-09-18T15:04:00Z", "kind": "placed"}, {"at": "2026-09-19T09:12:00Z", "kind": "fulfilled"}}})
	}
	if len(parts) == 3 && parts[2] == "shipments" && r.Method == http.MethodGet {
		if _, ok := d.lookupOrder(id); !ok {
			return problem(w, http.StatusNotFound, "order "+id+" is not on the desk")
		}
		return writeJSON(w, http.StatusOK, map[string]any{"data": []map[string]string{{"carrier": "usps", "tracking": "940011189922334" + id, "shippedAt": "2026-09-19T09:12:00Z"}}})
	}
	if len(parts) == 3 && parts[2] == "shipments" && r.Method == http.MethodPost {
		if _, ok := d.lookupOrder(id); !ok {
			return problem(w, http.StatusNotFound, "order "+id+" is not on the desk")
		}
		return writeJSON(w, http.StatusCreated, map[string]string{"carrier": "usps", "tracking": "pending"})
	}
	return problem(w, http.StatusNotFound, "no such order route")
}

func (d *Desk) customersRoute(w http.ResponseWriter, r *http.Request) int {
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")
	if path == "customers" && r.Method == http.MethodGet {
		d.mu.Lock()
		defer d.mu.Unlock()
		var data []*Customer
		for _, id := range []string{"cus_mara", "cus_jonas", "cus_priya"} {
			if c := d.customers[id]; c != nil {
				data = append(data, c)
			}
		}
		return writeJSON(w, http.StatusOK, map[string]any{"data": data})
	}
	if path == "customers" && r.Method == http.MethodPost {
		return writeJSON(w, http.StatusCreated, map[string]string{"id": "cus_new", "name": "New buyer", "email": "new@example.com", "since": "2026-10-01"})
	}
	if len(parts) < 2 || parts[0] != "customers" {
		return problem(w, http.StatusNotFound, "no such customer route")
	}
	id := parts[1]
	c, ok := d.lookupCustomer(id)
	if !ok {
		return problem(w, http.StatusNotFound, "customer "+id+" is not on the desk")
	}
	if len(parts) == 2 && r.Method == http.MethodGet {
		return writeJSON(w, http.StatusOK, c)
	}
	if len(parts) == 2 && r.Method == http.MethodPatch {
		return writeJSON(w, http.StatusOK, c)
	}
	if len(parts) == 3 && parts[2] == "addresses" && r.Method == http.MethodGet {
		return writeJSON(w, http.StatusOK, map[string]any{"data": []Address{c.Address}})
	}
	if len(parts) == 3 && parts[2] == "addresses" && r.Method == http.MethodPost {
		return writeJSON(w, http.StatusCreated, c.Address)
	}
	if len(parts) == 3 && parts[2] == "notes" && r.Method == http.MethodPost {
		return writeJSON(w, http.StatusCreated, map[string]string{"id": "cnote_" + id, "body": "noted"})
	}
	return problem(w, http.StatusNotFound, "no such customer route")
}

func (d *Desk) invoicesRoute(w http.ResponseWriter, r *http.Request) int {
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")
	if path == "invoices" && r.Method == http.MethodGet {
		d.mu.Lock()
		defer d.mu.Unlock()
		var data []*Invoice
		for _, id := range []string{"inv_2318", "inv_2304", "inv_2291"} {
			if inv := d.invoices[id]; inv != nil {
				data = append(data, inv)
			}
		}
		return writeJSON(w, http.StatusOK, map[string]any{"data": data})
	}
	if path == "invoices" && r.Method == http.MethodPost {
		return writeJSON(w, http.StatusCreated, map[string]any{"id": "inv_new", "status": "open", "currency": "USD", "totalAmount": 0})
	}
	if path == "payments" && r.Method == http.MethodGet {
		return writeJSON(w, http.StatusOK, map[string]any{"data": []map[string]any{payment()}})
	}
	if path == "payments" && r.Method == http.MethodPost {
		return writeJSON(w, http.StatusCreated, payment())
	}
	if len(parts) == 2 && parts[0] == "payments" && r.Method == http.MethodGet {
		if parts[1] != "pay_10482" {
			return problem(w, http.StatusNotFound, "payment "+parts[1]+" is not on the desk")
		}
		return writeJSON(w, http.StatusOK, payment())
	}
	if len(parts) < 2 || parts[0] != "invoices" {
		return problem(w, http.StatusNotFound, "no such billing route")
	}
	id := parts[1]
	inv, ok := d.lookupInvoice(id)
	if !ok {
		return problem(w, http.StatusNotFound, "invoice "+id+" is not on the desk")
	}
	if len(parts) == 2 && r.Method == http.MethodGet {
		return writeJSON(w, http.StatusOK, inv)
	}
	if len(parts) == 3 && parts[2] == "void" && r.Method == http.MethodPost {
		if inv.Status == "paid" {
			return problem(w, http.StatusConflict, "invoice "+id+" is already paid")
		}
		d.mu.Lock()
		inv.Status = "void"
		d.mu.Unlock()
		return writeJSON(w, http.StatusOK, inv)
	}
	if len(parts) == 3 && parts[2] == "lines" && r.Method == http.MethodGet {
		return writeJSON(w, http.StatusOK, map[string]any{"data": []map[string]any{
			{"sku": "WL-COAT-UM", "name": "Undyed wool coat", "quantity": 1, "amount": 42800},
			{"sku": "CD-TRAY-01", "name": "Cedar tray", "quantity": 2, "amount": 12800},
		}})
	}
	return problem(w, http.StatusNotFound, "no such billing route")
}

func payment() map[string]any {
	return map[string]any{
		"id": "pay_10482", "invoiceId": "inv_2291", "amount": 55600,
		"currency": "USD", "method": "card", "capturedAt": "2026-09-18T15:04:20Z",
	}
}

func (d *Desk) lookupOrder(id string) (*Order, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	o, ok := d.orders[id]
	return o, ok
}

func (d *Desk) lookupCustomer(id string) (*Customer, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	c, ok := d.customers[id]
	return c, ok
}

func (d *Desk) lookupInvoice(id string) (*Invoice, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	inv, ok := d.invoices[id]
	return inv, ok
}

func pageSize(raw string) int {
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return 20
	}
	return n
}

func writeJSON(w http.ResponseWriter, status int, body any) int {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
	return status
}

func problem(w http.ResponseWriter, status int, detail string) int {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"type": "about:blank", "title": http.StatusText(status), "status": status, "detail": detail,
	})
	return status
}
