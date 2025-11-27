package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/pressly/goose/v3"

	"github.com/llantera/hex/internal/adapters/http/handlers"
	"github.com/llantera/hex/internal/adapters/http/middleware"
	"github.com/llantera/hex/internal/adapters/http/router"
	"github.com/llantera/hex/internal/adapters/storage"
	"github.com/llantera/hex/internal/adapters/storage/postgres"
	addressapp "github.com/llantera/hex/internal/application/address"
	billingapp "github.com/llantera/hex/internal/application/billing"
	cartapp "github.com/llantera/hex/internal/application/cart"
	companyapp "github.com/llantera/hex/internal/application/company"
	customerrequestapp "github.com/llantera/hex/internal/application/customerrequest"
	notificationapp "github.com/llantera/hex/internal/application/notification"
	orderapp "github.com/llantera/hex/internal/application/order"
	pricelevelapp "github.com/llantera/hex/internal/application/pricelevel"
	tireapp "github.com/llantera/hex/internal/application/tire"
	userapp "github.com/llantera/hex/internal/application/user"
	"github.com/llantera/hex/internal/config"
	"github.com/llantera/hex/internal/platform/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	db, err := database.Open(cfg)
	if err != nil {
		log.Fatalf("database error: %v", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("configurar goose: %v", err)
	}
	if err := goose.Up(db, "migrations"); err != nil {
		log.Fatalf("migraciones pendientes: %v", err)
	}

	userRepo := postgres.NewUserRepository(db)
	companyRepo := postgres.NewCompanyRepository(db)
	tireRepo := postgres.NewTireRepository(db)
	brandRepo := postgres.NewTireBrandRepository(db)
	typeRepo := postgres.NewTireTypeRepository(db)
	inventoryRepo := postgres.NewInventoryRepository(db)
	priceRepo := postgres.NewPriceRepository(db)
	priceColumnRepo := postgres.NewPriceColumnRepository(db)
	customerRequestRepo := postgres.NewCustomerRequestRepository(db)
	priceLevelRepo := postgres.NewPriceLevelRepository(db)
	addressRepo := storage.NewAddressRepository(db)
	orderRepo := storage.NewOrderRepository(db)
	cartRepo := storage.NewCartRepository(db)
	billingRepo := storage.NewBillingRepository(db)
	notificationRepo := postgres.NewNotificationRepository(db)

	userService := userapp.NewService(userRepo)
	companyService := companyapp.NewService(companyRepo)
	priceLevelService := pricelevelapp.NewPriceLevelService(priceLevelRepo)
	tireService := tireapp.NewService(tireRepo, brandRepo, typeRepo, inventoryRepo, priceRepo, priceColumnRepo, priceLevelRepo)
	customerRequestService := customerrequestapp.NewService(customerRequestRepo)
	addressService := addressapp.NewService(addressRepo)
	orderService := orderapp.NewService(orderRepo, inventoryRepo)
	cartService := cartapp.NewService(cartRepo)
	billingService := billingapp.NewService(billingRepo)
	notificationService := notificationapp.NewService(notificationRepo)

	userHandler := handlers.NewUserHandler(userService)
	companyHandler := handlers.NewCompanyHandler(companyService)
	tireHandler := handlers.NewTireHandler(tireService)
	priceColumnHandler := handlers.NewPriceColumnHandler(tireService)
	priceLevelHandler := handlers.NewPriceLevelHandler(priceLevelService)
	brandHandler := handlers.NewTireBrandHandler(brandRepo)
	customerRequestHandler := handlers.NewCustomerRequestHandler(customerRequestService)
	authHandler := handlers.NewAuthHandler(userService, cfg.AuthSecret, cfg.AuthTokenTTL)
	addressHandler := handlers.NewAddressHandler(addressService)
	orderHandler := handlers.NewOrderHandler(orderService, notificationService, userRepo)
	cartHandler := handlers.NewCartHandler(cartService)
	billingHandler := handlers.NewBillingHandler(billingService)
	notificationHandler := handlers.NewNotificationHandler(notificationService)
	fileHandler := handlers.NewFileHandler(".")

	handler := router.New(router.Dependencies{
		UserHandler:            userHandler,
		CompanyHandler:         companyHandler,
		TireHandler:            tireHandler,
		TireBrandHandler:       brandHandler,
		PriceColumnHandler:     priceColumnHandler,
		CustomerRequestHandler: customerRequestHandler,
		PriceLevelHandler:      priceLevelHandler,
		AuthHandler:            authHandler,
		OrderHandler:           orderHandler,
		AddressHandler:         addressHandler,
		CartHandler:            cartHandler,
		BillingHandler:         billingHandler,
		NotificationHandler:    notificationHandler,
		FileHandler:            fileHandler,
		AuthSecret:             cfg.AuthSecret,
	})

	srv := &http.Server{
		Addr:    cfg.HTTPAddress(),
		Handler: middleware.WithCORS(handler),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server error: %v", err)
		}
	}()

	log.Printf("server listening on %s", srv.Addr)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ServerShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}
