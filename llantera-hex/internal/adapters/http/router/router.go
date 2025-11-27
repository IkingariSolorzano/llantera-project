package router

import (
	"net/http"
	"strings"

	"github.com/llantera/hex/internal/adapters/http/handlers"
	"github.com/llantera/hex/internal/adapters/http/middleware"
)

// Dependencies contiene los handlers que deben montarse en el router principal.
type Dependencies struct {
	UserHandler            *handlers.UserHandler
	CompanyHandler         *handlers.CompanyHandler
	TireHandler            *handlers.TireHandler
	TireBrandHandler       *handlers.TireBrandHandler
	PriceColumnHandler     *handlers.PriceColumnHandler
	CustomerRequestHandler *handlers.CustomerRequestHandler
	PriceLevelHandler      *handlers.PriceLevelHandler
	AuthHandler            *handlers.AuthHandler
	OrderHandler           *handlers.OrderHandler
	AddressHandler         *handlers.AddressHandler
	CartHandler            *handlers.CartHandler
	BillingHandler         *handlers.BillingHandler
	NotificationHandler    *handlers.NotificationHandler
	FileHandler            *handlers.FileHandler
	AuthSecret             string
}

// New construye el router principal de la API conectando los handlers recibidos.
func New(deps Dependencies) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	if deps.UserHandler != nil {
		var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isExactPath(r.URL.Path, "/api/users/") {
				deps.UserHandler.HandleCollection(w, r)
				return
			}
			deps.UserHandler.HandleResource(w, r)
		})
		if strings.TrimSpace(deps.AuthSecret) != "" {
			handler = middleware.WithAuth(deps.AuthSecret, middleware.WithRole([]string{"admin"}, handler))
		}
		mux.Handle("/api/users/", handler)
	}

	if deps.AuthHandler != nil {
		mux.HandleFunc("/api/auth/login", deps.AuthHandler.HandleLogin)
	}

	if deps.PriceColumnHandler != nil {
		var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isExactPath(r.URL.Path, "/api/price-columns/") {
				deps.PriceColumnHandler.HandleCollection(w, r)
				return
			}
			deps.PriceColumnHandler.HandleResource(w, r)
		})
		if strings.TrimSpace(deps.AuthSecret) != "" {
			handler = middleware.WithAuth(deps.AuthSecret, middleware.WithRole([]string{"admin", "employee"}, handler))
		}
		mux.Handle("/api/price-columns/", handler)
	}

	if deps.PriceLevelHandler != nil {
		var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isExactPath(r.URL.Path, "/api/price-levels/") {
				deps.PriceLevelHandler.HandleCollection(w, r)
				return
			}
			deps.PriceLevelHandler.HandleResource(w, r)
		})
		if strings.TrimSpace(deps.AuthSecret) != "" {
			handler = middleware.WithAuth(deps.AuthSecret, middleware.WithRole([]string{"admin", "employee"}, handler))
		}
		mux.Handle("/api/price-levels/", handler)
	}

	if deps.CompanyHandler != nil {
		var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isExactPath(r.URL.Path, "/api/companies/") {
				deps.CompanyHandler.HandleCollection(w, r)
				return
			}
			deps.CompanyHandler.HandleResource(w, r)
		})
		if strings.TrimSpace(deps.AuthSecret) != "" {
			handler = middleware.WithAuth(deps.AuthSecret, middleware.WithRole([]string{"admin", "employee"}, handler))
		}
		mux.Handle("/api/companies/", handler)
	}

	if deps.CustomerRequestHandler != nil {
		mux.HandleFunc("/api/customer-requests/", func(w http.ResponseWriter, r *http.Request) {
			// POST raíz público para "Quiero ser cliente" desde la landing.
			if isExactPath(r.URL.Path, "/api/customer-requests/") && r.Method == http.MethodPost {
				deps.CustomerRequestHandler.HandleCollection(w, r)
				return
			}

			// Resto de operaciones (listar, ver, actualizar, eliminar) requieren admin/employee.
			var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if isExactPath(r.URL.Path, "/api/customer-requests/") {
					deps.CustomerRequestHandler.HandleCollection(w, r)
					return
				}
				deps.CustomerRequestHandler.HandleResource(w, r)
			})
			if strings.TrimSpace(deps.AuthSecret) != "" {
				handler = middleware.WithAuth(deps.AuthSecret, middleware.WithRole([]string{"admin", "employee"}, handler))
			}
			handler.ServeHTTP(w, r)
		})
	}

	if deps.TireHandler != nil {
		// Importación de administración de llantas: /api/tires/admin/import
		var importHandler http.Handler = http.HandlerFunc(deps.TireHandler.HandleAdminImport)
		if strings.TrimSpace(deps.AuthSecret) != "" {
			importHandler = middleware.WithAuth(deps.AuthSecret, middleware.WithRole([]string{"admin"}, importHandler))
		}
		mux.Handle("/api/tires/admin/import", importHandler)

		// Exportación de administración de llantas: /api/tires/admin/export
		var exportHandler http.Handler = http.HandlerFunc(deps.TireHandler.HandleAdminExport)
		if strings.TrimSpace(deps.AuthSecret) != "" {
			exportHandler = middleware.WithAuth(deps.AuthSecret, middleware.WithRole([]string{"admin"}, exportHandler))
		}
		mux.Handle("/api/tires/admin/export", exportHandler)

		// Ruta de administración de llantas: /api/tires/admin/
		var adminHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isExactPath(r.URL.Path, "/api/tires/admin/") {
				deps.TireHandler.HandleAdmin(w, r)
				return
			}
			deps.TireHandler.HandleAdminResource(w, r)
		})
		if strings.TrimSpace(deps.AuthSecret) != "" {
			adminHandler = middleware.WithAuth(deps.AuthSecret, middleware.WithRole([]string{"admin"}, adminHandler))
		}
		mux.Handle("/api/tires/admin/", adminHandler)

		// Colección y recurso de llantas: /api/tires/
		var tiresHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isExactPath(r.URL.Path, "/api/tires/") {
				deps.TireHandler.HandleCollection(w, r)
				return
			}
			deps.TireHandler.HandleResource(w, r)
		})
		if strings.TrimSpace(deps.AuthSecret) != "" {
			tiresHandler = middleware.WithAuth(deps.AuthSecret, middleware.WithRole([]string{"admin"}, tiresHandler))
		}
		mux.Handle("/api/tires/", tiresHandler)

		// Catálogo público: se mantiene sin autenticación.
		mux.HandleFunc("/api/catalog/tires/", func(w http.ResponseWriter, r *http.Request) {
			if isExactPath(r.URL.Path, "/api/catalog/tires/") {
				deps.TireHandler.HandleCatalog(w, r)
				return
			}
			http.NotFound(w, r)
		})
	}

	if deps.TireBrandHandler != nil {
		var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isExactPath(r.URL.Path, "/api/brands/") {
				deps.TireBrandHandler.HandleCollection(w, r)
				return
			}
			deps.TireBrandHandler.HandleResource(w, r)
		})
		if strings.TrimSpace(deps.AuthSecret) != "" {
			handler = middleware.WithAuth(deps.AuthSecret, middleware.WithRole([]string{"admin", "employee"}, handler))
		}
		mux.Handle("/api/brands/", handler)
	}

	// Rutas de direcciones de cliente (requiere autenticación de customer)
	if deps.AddressHandler != nil {
		var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isExactPath(r.URL.Path, "/api/addresses/") {
				deps.AddressHandler.HandleCollection(w, r)
				return
			}
			deps.AddressHandler.HandleResource(w, r)
		})
		if strings.TrimSpace(deps.AuthSecret) != "" {
			handler = middleware.WithAuth(deps.AuthSecret, handler)
		}
		mux.Handle("/api/addresses/", handler)
	}

	// Rutas de pedidos de cliente
	if deps.OrderHandler != nil {
		// Pedidos del cliente autenticado
		var customerOrdersHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isExactPath(r.URL.Path, "/api/orders/") {
				deps.OrderHandler.HandleCustomerOrders(w, r)
				return
			}
			deps.OrderHandler.HandleCustomerOrderResource(w, r)
		})
		if strings.TrimSpace(deps.AuthSecret) != "" {
			customerOrdersHandler = middleware.WithAuth(deps.AuthSecret, customerOrdersHandler)
		}
		mux.Handle("/api/orders/", customerOrdersHandler)

		// Pedidos para administración (admin/employee)
		var adminOrdersHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isExactPath(r.URL.Path, "/api/admin/orders/") {
				deps.OrderHandler.HandleAdminOrders(w, r)
				return
			}
			// Subida de facturas: /api/admin/orders/{id}/invoice
			if strings.Contains(r.URL.Path, "/invoice") {
				deps.OrderHandler.HandleUploadInvoice(w, r)
				return
			}
			deps.OrderHandler.HandleAdminOrderResource(w, r)
		})
		if strings.TrimSpace(deps.AuthSecret) != "" {
			adminOrdersHandler = middleware.WithAuth(deps.AuthSecret, middleware.WithRole([]string{"admin", "employee"}, adminOrdersHandler))
		}
		mux.Handle("/api/admin/orders/", adminOrdersHandler)
	}

	// Rutas del carrito de compras (requiere autenticación)
	if deps.CartHandler != nil {
		// Carrito: GET, POST, DELETE sobre /api/cart/
		var cartHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isExactPath(r.URL.Path, "/api/cart/") {
				deps.CartHandler.HandleCollection(w, r)
				return
			}
			deps.CartHandler.HandleResource(w, r)
		})
		if strings.TrimSpace(deps.AuthSecret) != "" {
			cartHandler = middleware.WithAuth(deps.AuthSecret, cartHandler)
		}
		mux.Handle("/api/cart/", cartHandler)
	}

	// Rutas de datos de facturación de cliente (requiere autenticación)
	if deps.BillingHandler != nil {
		var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isExactPath(r.URL.Path, "/api/billing/") {
				deps.BillingHandler.HandleCollection(w, r)
				return
			}
			deps.BillingHandler.HandleResource(w, r)
		})
		if strings.TrimSpace(deps.AuthSecret) != "" {
			handler = middleware.WithAuth(deps.AuthSecret, handler)
		}
		mux.Handle("/api/billing/", handler)
	}

	// Rutas de notificaciones (requiere autenticación)
	if deps.NotificationHandler != nil {
		var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isExactPath(r.URL.Path, "/api/notifications/") {
				deps.NotificationHandler.HandleCollection(w, r)
				return
			}
			deps.NotificationHandler.HandleResource(w, r)
		})
		if strings.TrimSpace(deps.AuthSecret) != "" {
			handler = middleware.WithAuth(deps.AuthSecret, handler)
		}
		mux.Handle("/api/notifications/", handler)
	}

	// Rutas para servir archivos (facturas, etc.) - requiere autenticación
	if deps.FileHandler != nil {
		var handler http.Handler = http.HandlerFunc(deps.FileHandler.ServeFile)
		if strings.TrimSpace(deps.AuthSecret) != "" {
			handler = middleware.WithAuth(deps.AuthSecret, handler)
		}
		mux.Handle("/api/files/", handler)
	}

	return mux
}

func isExactPath(got, expected string) bool {
	trimmed := strings.TrimSuffix(got, "/")
	if trimmed == "" {
		trimmed = "/"
	}
	expected = strings.TrimSuffix(expected, "/")
	if expected == "" {
		expected = "/"
	}
	return trimmed == expected
}
