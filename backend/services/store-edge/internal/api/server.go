package api

import (
	"context"
	"errors"
	"net/http"
	"os"
	"time"

	"mercadia.dev/pos/platform/httpapi"
	platformmigrate "mercadia.dev/pos/platform/migrate"
	"mercadia.dev/pos/services/store-edge/internal/app"
	"mercadia.dev/pos/services/store-edge/internal/domain"
	centralclient "mercadia.dev/pos/services/store-edge/internal/infra/central"
	haclient "mercadia.dev/pos/services/store-edge/internal/infra/hardwareagent"
	"mercadia.dev/pos/services/store-edge/internal/infra/memory"
	"mercadia.dev/pos/services/store-edge/internal/infra/postgres"
)

const version = "0.1.0"

type StatusResponse struct {
	StoreID      string    `json:"storeId"`
	Mode         string    `json:"mode"`
	BusinessDate string    `json:"businessDate"`
	GeneratedAt  time.Time `json:"generatedAt"`
}

type OpenOperationalDayRequest struct {
	StoreID      string `json:"storeId"`
	BusinessDate string `json:"businessDate"`
	OpenedByID   string `json:"openedById"`
}

type CloseOperationalDayRequest struct {
	ClosedByID      string `json:"closedById"`
	OverrideNoSales bool   `json:"overrideNoSales,omitempty"`
	OverrideActorID string `json:"overrideActorId,omitempty"`
}

type OperationalDayAcceptedResponse struct {
	OperationalDay OperationalDayResponse `json:"operationalDay"`
}

type OperationalDayCloseReadinessResponse struct {
	OperationalDay OperationalDayResponse  `json:"operationalDay"`
	CanClose       bool                    `json:"canClose"`
	Blockers       []OperationalDayBlocker `json:"blockers"`
}

type OperationalDaySummaryResponse struct {
	OperationalDay OperationalDayResponse       `json:"operationalDay"`
	CanClose       bool                         `json:"canClose"`
	Blockers       []OperationalDayBlocker      `json:"blockers"`
	Shifts         OperationalDayShiftSummary   `json:"shifts"`
	Cash           OperationalDayCashSummary    `json:"cash"`
	Receipts       OperationalDayReceiptSummary `json:"receipts"`
	Payments       OperationalDayPaymentSummary `json:"payments"`
	Fiscal         OperationalDayFiscalSummary  `json:"fiscal"`
}

type OperationalDayShiftSummary struct {
	TotalCount  int `json:"totalCount"`
	OpenCount   int `json:"openCount"`
	ClosedCount int `json:"closedCount"`
}

type OperationalDayReceiptSummary struct {
	TotalCount           int   `json:"totalCount"`
	DraftCount           int   `json:"draftCount"`
	PaymentStartedCount  int   `json:"paymentStartedCount"`
	PaidCount            int   `json:"paidCount"`
	FiscalizedCount      int   `json:"fiscalizedCount"`
	CancelledCount       int   `json:"cancelledCount"`
	UnresolvedCount      int   `json:"unresolvedCount"`
	FiscalizedSalesMinor int64 `json:"fiscalizedSalesMinor"`
}

type OperationalDayPaymentSummary struct {
	TotalCount         int                           `json:"totalCount"`
	CapturedCount      int                           `json:"capturedCount"`
	CapturedTotalMinor int64                         `json:"capturedTotalMinor"`
	Methods            []OperationalDayPaymentMethod `json:"methods"`
}

type OperationalDayPaymentMethod struct {
	Method             domain.PaymentMethod `json:"method"`
	CapturedCount      int                  `json:"capturedCount"`
	CapturedTotalMinor int64                `json:"capturedTotalMinor"`
}

type OperationalDayFiscalSummary struct {
	TotalCount           int   `json:"totalCount"`
	FiscalizedCount      int   `json:"fiscalizedCount"`
	FiscalizedTotalMinor int64 `json:"fiscalizedTotalMinor"`
}

type OperationalDayCashSummary struct {
	Balances           []CashBalanceResponse `json:"balances"`
	NonZeroDrawerCount int                   `json:"nonZeroDrawerCount"`
	Recounts           CashRecountSummary    `json:"recounts"`
}

type CashRecountSummary struct {
	TotalCount               int `json:"totalCount"`
	BalancedCount            int `json:"balancedCount"`
	DiscrepancyCount         int `json:"discrepancyCount"`
	OpenDiscrepancyCount     int `json:"openDiscrepancyCount"`
	ResolvedDiscrepancyCount int `json:"resolvedDiscrepancyCount"`
}

type OperationalDayResponse struct {
	ID           string                      `json:"id"`
	StoreID      string                      `json:"storeId"`
	BusinessDate string                      `json:"businessDate"`
	Status       domain.OperationalDayStatus `json:"status"`
	OpenedByID   string                      `json:"openedById"`
	ClosedByID   string                      `json:"closedById,omitempty"`
	OpenedAt     time.Time                   `json:"openedAt"`
	ClosedAt     time.Time                   `json:"closedAt,omitempty"`
	UpdatedAt    time.Time                   `json:"updatedAt"`
}

type OperationalDayBlocker struct {
	Code        string                               `json:"code"`
	Severity    domain.OperationalDayBlockerSeverity `json:"severity"`
	Message     string                               `json:"message"`
	ReferenceID string                               `json:"referenceId,omitempty"`
}

type ReceiptAcceptedResponse struct {
	Receipt ReceiptResponse `json:"receipt"`
}

type ReceiptsResponse struct {
	Receipts []ReceiptResponse `json:"receipts"`
}

type HeartbeatResponse struct {
	Terminal TerminalResponse `json:"terminal"`
}

type TerminalHeartbeatRequest struct {
	StoreID         string              `json:"storeId"`
	Kind            domain.TerminalKind `json:"kind"`
	SoftwareVersion string              `json:"softwareVersion,omitempty"`
}

type OpenReceiptRequest struct {
	StoreID    string `json:"storeId"`
	TerminalID string `json:"terminalId"`
	CashierID  string `json:"cashierId"`
	Channel    string `json:"channel,omitempty"`
}

type AddReceiptLineRequest struct {
	ProductID      string `json:"productId"`
	Barcode        string `json:"barcode,omitempty"`
	Name           string `json:"name"`
	Quantity       int64  `json:"quantity"`
	UnitPriceMinor int64  `json:"unitPriceMinor"`
}

type ScanReceiptLineRequest struct {
	Barcode  string `json:"barcode"`
	Quantity int64  `json:"quantity"`
}

type CancelReceiptRequest struct {
	Reason       string `json:"reason"`
	ActorID      string `json:"actorId"`
	ApprovedByID string `json:"approvedById,omitempty"`
}

type ProductResponse struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Barcodes       []string `json:"barcodes"`
	UnitPriceMinor int64    `json:"unitPriceMinor"`
	TaxCategoryID  string   `json:"taxCategoryId,omitempty"`
	Active         bool     `json:"active"`
}

type CreatePaymentRequest struct {
	Method            domain.PaymentMethod `json:"method"`
	AmountMinor       int64                `json:"amountMinor"`
	ProviderReference string               `json:"providerReference,omitempty"`
}

type CancelPaymentRequest struct {
	ActorID string `json:"actorId,omitempty"`
	Reason  string `json:"reason,omitempty"`
}

type RefundPaymentRequest struct {
	ActorID string `json:"actorId,omitempty"`
	Reason  string `json:"reason,omitempty"`
}

type PaymentAcceptedResponse struct {
	Payment PaymentResponse `json:"payment"`
}

type PaymentsResponse struct {
	Payments []PaymentResponse `json:"payments"`
}

type PaymentResponse struct {
	ID                string               `json:"id"`
	ReceiptID         string               `json:"receiptId"`
	Method            domain.PaymentMethod `json:"method"`
	Status            domain.PaymentStatus `json:"status"`
	AmountMinor       int64                `json:"amountMinor"`
	ProviderReference string               `json:"providerReference,omitempty"`
	CreatedAt         time.Time            `json:"createdAt"`
	UpdatedAt         time.Time            `json:"updatedAt"`
	CapturedAt        time.Time            `json:"capturedAt"`
}

type CreateFiscalDocumentRequest struct {
	DeviceID string `json:"deviceId"`
}

type FiscalDocumentAcceptedResponse struct {
	Document FiscalDocumentResponse `json:"document"`
}

type FiscalDocumentsResponse struct {
	Documents []FiscalDocumentResponse `json:"documents"`
}

type FiscalDocumentResponse struct {
	ID           string                      `json:"id"`
	ReceiptID    string                      `json:"receiptId"`
	Kind         domain.FiscalDocumentKind   `json:"kind"`
	Status       domain.FiscalDocumentStatus `json:"status"`
	AmountMinor  int64                       `json:"amountMinor"`
	DeviceID     string                      `json:"deviceId"`
	FiscalSign   string                      `json:"fiscalSign"`
	FiscalizedAt time.Time                   `json:"fiscalizedAt"`
	CreatedAt    time.Time                   `json:"createdAt"`
}

type CreateCashMovementRequest struct {
	Type              domain.CashMovementType  `json:"type"`
	FromContainerID   string                   `json:"fromContainerId"`
	FromContainerType domain.CashContainerType `json:"fromContainerType"`
	ToContainerID     string                   `json:"toContainerId"`
	ToContainerType   domain.CashContainerType `json:"toContainerType"`
	AmountMinor       int64                    `json:"amountMinor"`
	Currency          string                   `json:"currency,omitempty"`
	Reason            string                   `json:"reason,omitempty"`
	ActorID           string                   `json:"actorId"`
	ApprovedByID      string                   `json:"approvedById,omitempty"`
}

type CashMovementAcceptedResponse struct {
	Movement CashMovementResponse `json:"movement"`
}

type CashMovementsResponse struct {
	Movements []CashMovementResponse `json:"movements"`
}

type CashBalancesResponse struct {
	Balances []CashBalanceResponse `json:"balances"`
}

type CreateCashRecountRequest struct {
	ContainerID   string                   `json:"containerId"`
	ContainerType domain.CashContainerType `json:"containerType"`
	Currency      string                   `json:"currency,omitempty"`
	CountedMinor  int64                    `json:"countedMinor"`
	Reason        string                   `json:"reason,omitempty"`
	ActorID       string                   `json:"actorId"`
	ApprovedByID  string                   `json:"approvedById,omitempty"`
}

type CashRecountAcceptedResponse struct {
	Recount CashRecountResponse `json:"recount"`
}

type ResolveCashRecountRequest struct {
	ResolutionNote string `json:"resolutionNote"`
	ActorID        string `json:"actorId"`
	ApprovedByID   string `json:"approvedById"`
}

type CashRecountsResponse struct {
	Recounts []CashRecountResponse `json:"recounts"`
}

type CashMovementResponse struct {
	ID                string                    `json:"id"`
	StoreID           string                    `json:"storeId"`
	Type              domain.CashMovementType   `json:"type"`
	FromContainerID   string                    `json:"fromContainerId"`
	FromContainerType domain.CashContainerType  `json:"fromContainerType"`
	ToContainerID     string                    `json:"toContainerId"`
	ToContainerType   domain.CashContainerType  `json:"toContainerType"`
	AmountMinor       int64                     `json:"amountMinor"`
	Currency          string                    `json:"currency"`
	Reason            string                    `json:"reason,omitempty"`
	ActorID           string                    `json:"actorId"`
	ApprovedByID      string                    `json:"approvedById,omitempty"`
	Status            domain.CashMovementStatus `json:"status"`
	CreatedAt         time.Time                 `json:"createdAt"`
}

type CashBalanceResponse struct {
	StoreID        string                   `json:"storeId"`
	ContainerID    string                   `json:"containerId"`
	ContainerType  domain.CashContainerType `json:"containerType"`
	Currency       string                   `json:"currency"`
	BalanceMinor   int64                    `json:"balanceMinor"`
	LastMovementAt time.Time                `json:"lastMovementAt"`
}

type CashRecountResponse struct {
	ID               string                             `json:"id"`
	StoreID          string                             `json:"storeId"`
	BusinessDate     string                             `json:"businessDate,omitempty"`
	ContainerID      string                             `json:"containerId"`
	ContainerType    domain.CashContainerType           `json:"containerType"`
	Currency         string                             `json:"currency"`
	ExpectedMinor    int64                              `json:"expectedMinor"`
	CountedMinor     int64                              `json:"countedMinor"`
	DiscrepancyMinor int64                              `json:"discrepancyMinor"`
	Reason           string                             `json:"reason,omitempty"`
	ActorID          string                             `json:"actorId"`
	ApprovedByID     string                             `json:"approvedById,omitempty"`
	Status           domain.CashRecountStatus           `json:"status"`
	ResolutionStatus domain.CashRecountResolutionStatus `json:"resolutionStatus"`
	ResolutionNote   string                             `json:"resolutionNote,omitempty"`
	ResolvedByID     string                             `json:"resolvedById,omitempty"`
	ResolvedAt       time.Time                          `json:"resolvedAt,omitempty"`
	CreatedAt        time.Time                          `json:"createdAt"`
}

type OpenShiftRequest struct {
	StoreID          string `json:"storeId"`
	TerminalID       string `json:"terminalId"`
	CashierID        string `json:"cashierId"`
	DrawerID         string `json:"drawerId"`
	OpeningCashMinor int64  `json:"openingCashMinor"`
}

type CloseShiftRequest struct {
	ClosingCashMinor int64  `json:"closingCashMinor"`
	SafeID           string `json:"safeId,omitempty"`
	ActorID          string `json:"actorId,omitempty"`
	ApprovedByID     string `json:"approvedById,omitempty"`
}

type ShiftAcceptedResponse struct {
	Shift ShiftResponse `json:"shift"`
}

type ShiftsResponse struct {
	Shifts []ShiftResponse `json:"shifts"`
}

type ShiftResponse struct {
	ID               string             `json:"id"`
	StoreID          string             `json:"storeId"`
	OperationalDayID string             `json:"operationalDayId,omitempty"`
	BusinessDate     string             `json:"businessDate,omitempty"`
	TerminalID       string             `json:"terminalId"`
	CashierID        string             `json:"cashierId"`
	DrawerID         string             `json:"drawerId"`
	Status           domain.ShiftStatus `json:"status"`
	OpeningCashMinor int64              `json:"openingCashMinor"`
	ClosingCashMinor int64              `json:"closingCashMinor"`
	OpenedAt         time.Time          `json:"openedAt"`
	ClosedAt         time.Time          `json:"closedAt,omitempty"`
	UpdatedAt        time.Time          `json:"updatedAt"`
}

type ReceiptResponse struct {
	ID                 string                `json:"id"`
	StoreID            string                `json:"storeId"`
	OperationalDayID   string                `json:"operationalDayId,omitempty"`
	BusinessDate       string                `json:"businessDate,omitempty"`
	ShiftID            string                `json:"shiftId,omitempty"`
	TerminalID         string                `json:"terminalId"`
	CashierID          string                `json:"cashierId"`
	DrawerID           string                `json:"drawerId,omitempty"`
	Channel            string                `json:"channel"`
	Status             domain.ReceiptStatus  `json:"status"`
	Lines              []ReceiptLineResponse `json:"lines"`
	CancelReason       string                `json:"cancelReason,omitempty"`
	CancelledByID      string                `json:"cancelledById,omitempty"`
	CancelApprovedByID string                `json:"cancelApprovedById,omitempty"`
	CancelledAt        time.Time             `json:"cancelledAt,omitempty"`
	TotalMinor         int64                 `json:"totalMinor"`
	CreatedAt          time.Time             `json:"createdAt"`
	UpdatedAt          time.Time             `json:"updatedAt"`
}

type ReceiptLineResponse struct {
	ID                  string    `json:"id"`
	ProductID           string    `json:"productId"`
	Barcode             string    `json:"barcode,omitempty"`
	Name                string    `json:"name"`
	Quantity            int64     `json:"quantity"`
	UnitPriceMinor      int64     `json:"unitPriceMinor"`
	DiscountMinor       int64     `json:"discountMinor,omitempty"`
	DiscountReason      string    `json:"discountReason,omitempty"`
	DiscountAppliedByID string    `json:"discountAppliedById,omitempty"`
	TotalMinor          int64     `json:"totalMinor"`
	AddedAt             time.Time `json:"addedAt"`
}

type PaginatedTerminalsResponse struct {
	Items      []TerminalResponse `json:"items"`
	TotalCount int                `json:"totalCount"`
}

type TerminalResponse struct {
	ID              string                `json:"id"`
	StoreID         string                `json:"storeId"`
	Kind            domain.TerminalKind   `json:"kind"`
	Status          domain.TerminalStatus `json:"status"`
	SoftwareVersion string                `json:"softwareVersion,omitempty"`
	LastSeenAt      time.Time             `json:"lastSeenAt"`
	UpdatedAt       time.Time             `json:"updatedAt"`
}

type ServerOptions struct {
	DatabaseURL           string
	MigrationsDir         string
	CentralBackendURL     string
	HardwareAgentURL      string
	UseHardwareAgent      bool
	HardwareAgentFallback bool
	ReadinessChecks       []func(context.Context) error
	BrokerConnected       func() bool
	DefaultStoreID        string
}

type ServerBundle struct {
	Handler     http.Handler
	Outbox      app.OutboxRepository
	CatalogSync *app.CatalogSyncService
	DefaultStoreID string
}

func NewServer() http.Handler {
	mux, _, _, _, _, err := buildServer(ServerOptions{})
	if err != nil {
		panic(err)
	}
	return mux
}

func NewServerWithOptions(opts ServerOptions) (http.Handler, error) {
	bundle, err := NewServerBundle(opts)
	if err != nil {
		return nil, err
	}
	return bundle.Handler, nil
}

func NewServerBundle(opts ServerOptions) (*ServerBundle, error) {
	mux, _, outbox, catalogSync, defaultStoreID, err := buildServer(opts)
	if err != nil {
		return nil, err
	}
	return &ServerBundle{
		Handler:        mux,
		Outbox:         outbox,
		CatalogSync:    catalogSync,
		DefaultStoreID: defaultStoreID,
	}, nil
}

func OpenAPI() map[string]any {
	_, spec, _, _, _, err := buildServer(ServerOptions{})
	if err != nil {
		panic(err)
	}
	return spec.OpenAPI()
}

func buildServer(opts ServerOptions) (*http.ServeMux, *httpapi.Spec, app.OutboxRepository, *app.CatalogSyncService, string, error) {
	databaseURL := opts.DatabaseURL
	if databaseURL == "" {
		databaseURL = os.Getenv("MERCADIA_STORE_EDGE_DATABASE_URL")
	}

	migrationsDir := opts.MigrationsDir
	if migrationsDir == "" {
		migrationsDir = postgres.DefaultMigrationsDir()
	}

	centralBackendURL := opts.CentralBackendURL
	if centralBackendURL == "" {
		centralBackendURL = os.Getenv("MERCADIA_CENTRAL_BACKEND_URL")
	}
	if centralBackendURL == "" {
		centralBackendURL = centralclient.DefaultBaseURL()
	}

	hardwareAgentURL := opts.HardwareAgentURL
	if hardwareAgentURL == "" {
		hardwareAgentURL = os.Getenv("MERCADIA_HARDWARE_AGENT_URL")
	}
	if hardwareAgentURL == "" {
		hardwareAgentURL = haclient.DefaultBaseURL()
	}

	useHardwareAgent := opts.UseHardwareAgent
	if !useHardwareAgent {
		useHardwareAgent = os.Getenv("MERCADIA_STORE_EDGE_USE_HARDWARE_AGENT") == "true"
	}

	hardwareAgentFallback := opts.HardwareAgentFallback
	if !hardwareAgentFallback {
		hardwareAgentFallback = os.Getenv("MERCADIA_STORE_EDGE_HARDWARE_AGENT_FALLBACK") != "false"
	}

	defaultStoreID := opts.DefaultStoreID
	if defaultStoreID == "" {
		defaultStoreID = os.Getenv("MERCADIA_STORE_EDGE_DEFAULT_STORE_ID")
	}
	if defaultStoreID == "" {
		defaultStoreID = "store-1"
	}

	terminalOfflineAfter := 60 * time.Second
	if raw := os.Getenv("MERCADIA_STORE_EDGE_TERMINAL_OFFLINE_AFTER"); raw != "" {
		if parsed, err := time.ParseDuration(raw); err == nil {
			terminalOfflineAfter = parsed
		}
	}

	readinessChecks := append([]func(context.Context) error(nil), opts.ReadinessChecks...)

	var store storeRepositories

	if databaseURL != "" {
		ctx := context.Background()
		pgStore, err := postgres.NewStore(ctx, databaseURL, postgres.WithProducts(demoProducts()...))
		if err != nil {
			return nil, nil, nil, nil, "", err
		}
		migrationResult, err := postgres.RunMigrations(ctx, pgStore.Pool(), migrationsDir)
		if err != nil {
			platformmigrate.LogError("store-edge", migrationsDir, err)
			pgStore.Close()
			return nil, nil, nil, nil, "", err
		}
		platformmigrate.LogResult(migrationResult)
		if err := pgStore.SeedDemoActors(ctx); err != nil {
			pgStore.Close()
			return nil, nil, nil, nil, "", err
		}
		readinessChecks = append(readinessChecks, pgStore.Ping)
		store = pgStore
	} else {
		store = memory.NewStore(memory.WithProducts(demoProducts()...), memory.WithDemoActors())
	}

	var catalogSync *app.CatalogSyncService
	if centralBackendURL != "" {
		centralClient := centralclient.NewClient(centralBackendURL, nil)
		catalogSync = app.NewCatalogSyncService(store, store, centralClient)
	}

	var hardwareAgent *haclient.Client
	if useHardwareAgent {
		hardwareAgent = haclient.NewClient(hardwareAgentURL, nil)
	}

	var systemOptions []httpapi.SystemRoutesOption
	if len(readinessChecks) > 0 {
		systemOptions = append(systemOptions, httpapi.WithReadinessCheck(combineReadinessChecks(readinessChecks)))
	}

	mux, spec := wireServer(wireConfig{
		store:                 store,
		brokerConnected:       opts.BrokerConnected,
		catalogSync:           catalogSync,
		hardwareAgent:         hardwareAgent,
		useHardwareAgent:      useHardwareAgent,
		hardwareAgentFallback: hardwareAgentFallback,
		terminalOfflineAfter:  terminalOfflineAfter,
	}, systemOptions...)
	return mux, spec, store, catalogSync, defaultStoreID, nil
}

func combineReadinessChecks(checks []func(context.Context) error) func(context.Context) error {
	return func(ctx context.Context) error {
		for _, check := range checks {
			if check == nil {
				continue
			}
			if err := check(ctx); err != nil {
				return err
			}
		}
		return nil
	}
}

type storeRepositories interface {
	app.ReceiptRepository
	app.IdempotencyStore
	app.ProductRepository
	app.CatalogSyncStateRepository
	app.PaymentRepository
	app.FiscalRepository
	app.CashRepository
	app.ShiftRepository
	app.OperationalDayRepository
	app.TerminalRepository
	app.ShiftReceiptRepository
	app.OperationalDayShiftRepository
	app.OperationalDayReceiptRepository
	app.OperationalDayCashRepository
	app.OutboxRepository
	app.ActorRepository
	app.SessionRepository
	app.ReturnRepository
	app.OperationJournalRepository
}

type wireConfig struct {
	store                 storeRepositories
	brokerConnected       func() bool
	catalogSync           *app.CatalogSyncService
	hardwareAgent         *haclient.Client
	useHardwareAgent      bool
	hardwareAgentFallback bool
	terminalOfflineAfter  time.Duration
}

func wireServer(config wireConfig, systemOptions ...httpapi.SystemRoutesOption) (*http.ServeMux, *httpapi.Spec) {
	store := config.store
	outbox := app.NewOutboxService(store)
	journal := app.NewOperationJournalService(store)
	auth := app.NewAuthService(store, store)
	terminalEvents := app.NewTerminalEventHub()

	operationalDays := app.NewOperationalDayService(store, store, store, store, store, app.WithOperationalDayOutboxRecorder(outbox))
	checkout := app.NewCheckoutService(store, store, app.WithProductRepository(store), app.WithStoreOperations(store, store))
	catalog := app.NewCatalogService(store)

	paymentOptions := []app.PaymentOption{
		app.WithPaymentCashLedger(store),
		app.WithPaymentOutboxRecorder(outbox),
	}
	fiscalOptions := []app.FiscalizationOption{
		app.WithFiscalizationOutboxRecorder(outbox),
	}
	if config.useHardwareAgent && config.hardwareAgent != nil {
		paymentOptions = append(paymentOptions, app.WithCardPaymentTerminal(config.hardwareAgent, "sim-payment-1", config.hardwareAgentFallback))
		fiscalOptions = append(fiscalOptions, app.WithFiscalReceiptPrinter(config.hardwareAgent, config.hardwareAgentFallback))
	}

	payments := app.NewPaymentService(store, store, store, paymentOptions...)
	fiscalization := app.NewFiscalizationService(store, store, store, store, fiscalOptions...)
	cash := app.NewCashService(store, store, app.WithCashOutboxRecorder(outbox), app.WithCashJournal(journal))
	shifts := app.NewShiftService(store, store, app.WithShiftCashLedger(store), app.WithShiftReceiptRepository(store), app.WithShiftOperationalDayRepository(store))
	terminalOptions := []app.TerminalOption{app.WithTerminalEventPublisher(terminalEvents)}
	if config.terminalOfflineAfter > 0 {
		terminalOptions = append(terminalOptions, app.WithTerminalOfflineAfter(config.terminalOfflineAfter))
	}
	terminals := app.NewTerminalService(store, store, terminalOptions...)
	monitoringOptions := []app.TerminalMonitoringOption{}
	if config.terminalOfflineAfter > 0 {
		monitoringOptions = append(monitoringOptions, app.WithTerminalMonitoringOfflineAfter(config.terminalOfflineAfter))
	}
	terminalMonitoring := app.NewTerminalMonitoringService(store, store, store, store, cash, monitoringOptions...)
	returns := app.NewReturnsService(store, store, store, auth, app.WithReturnsJournal(journal))
	returnSettlement := app.NewReturnSettlementService(store, store, store, payments, store,
		app.WithReturnSettlementOutboxRecorder(outbox),
		app.WithReturnSettlementJournal(journal),
	)
	discounts := app.NewDiscountService(store, store, auth, app.WithDiscountJournal(journal))
	marking := app.NewMarkingService(store)

	info := httpapi.ServiceInfo{
		Name:        "store-edge",
		Title:       "Mercadia Store Edge",
		Description: "Store-local operational API for POS, SCO/KSO, senior cashier, assistant, and store admin clients.",
		Version:     version,
	}

	mux := http.NewServeMux()
	spec := httpapi.NewSpec(info)
	httpapi.MountSystemRoutes(mux, spec, info, systemOptions...)
	mountRoutes(mux, spec, outbox, config.brokerConnected, operationalDays, checkout, catalog, payments, fiscalization, cash, shifts, terminals)
	mountMonitoringRoutes(mux, spec, terminalMonitoring)
	mountDomainRoutes(mux, spec, auth, returns, returnSettlement, discounts, marking, journal)
	mountCatalogSyncRoute(mux, spec, config.catalogSync)
	mountTerminalEventsRoute(mux, terminalEvents)

	return mux, spec
}

type OutboxStatusResponse struct {
	PendingCount    int64 `json:"pendingCount"`
	PublishedCount  int64 `json:"publishedCount"`
	BrokerConnected bool  `json:"brokerConnected"`
}

func mountRoutes(mux *http.ServeMux, spec *httpapi.Spec, outbox *app.OutboxService, brokerConnected func() bool, operationalDays *app.OperationalDayService, checkout *app.CheckoutService, catalog *app.CatalogService, payments *app.PaymentService, fiscalization *app.FiscalizationService, cash *app.CashService, shifts *app.ShiftService, terminals *app.TerminalService) {
	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/store-edge/sync/outbox-status",
		OperationID: "getOutboxStatus",
		Summary:     "Get outbox synchronization status",
		Tags:        []string{"sync"},
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Outbox status", Schema: outboxStatusResponseSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		connected := false
		if brokerConnected != nil {
			connected = brokerConnected()
		}
		status, err := outbox.Status(r.Context(), connected)
		if err != nil {
			httpapi.WriteProblem(w, http.StatusInternalServerError, "outbox_status_failed", "Failed to read outbox status", err.Error())
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, OutboxStatusResponse{
			PendingCount:    status.PendingCount,
			PublishedCount:  status.PublishedCount,
			BrokerConnected: status.BrokerConnected,
		})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/store-edge/status",
		OperationID: "getStoreEdgeStatus",
		Summary:     "Get Store Edge operational status",
		Tags:        []string{"store-operations"},
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Store Edge status", Schema: statusResponseSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		httpapi.WriteJSON(w, http.StatusOK, StatusResponse{
			StoreID:      "local-store",
			Mode:         "online",
			BusinessDate: time.Now().UTC().Format("2006-01-02"),
			GeneratedAt:  time.Now().UTC(),
		})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:              http.MethodPost,
		Path:                "/v1/operational-days",
		OperationID:         "openOperationalDay",
		Summary:             "Open operational day",
		Tags:                []string{"store-operations"},
		RequiresIdempotency: true,
		RequestBody: &httpapi.BodySpec{
			Description: "Operational day opening command",
			Required:    true,
			Schema:      openOperationalDayRequestSchema(),
		},
		Responses: map[string]httpapi.ResponseSpec{
			"202": {Description: "Operational day opened", Schema: operationalDayAcceptedResponseSchema()},
			"400": {Description: "Invalid operational day command", Schema: httpapi.ProblemSchema()},
			"409": {Description: "Operational day or idempotency conflict", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		if _, err := httpapi.RequireIdempotencyKey(r); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "idempotency_key_required", "Idempotency key is required", err.Error())
			return
		}
		var request OpenOperationalDayRequest
		if err := httpapi.DecodeJSON(r, &request); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_json", "Invalid JSON", err.Error())
			return
		}
		result, err := operationalDays.OpenOperationalDay(r.Context(), app.OpenOperationalDayCommand{
			IdempotencyKey: r.Header.Get("Idempotency-Key"),
			StoreID:        request.StoreID,
			BusinessDate:   request.BusinessDate,
			OpenedByID:     request.OpenedByID,
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusAccepted, OperationalDayAcceptedResponse{
			OperationalDay: operationalDayResponse(result.Day),
		})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/operational-days/{operationalDayId}",
		OperationID: "getOperationalDay",
		Summary:     "Get operational day",
		Tags:        []string{"store-operations"},
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Operational day", Schema: operationalDayResponseSchema()},
			"404": {Description: "Operational day was not found", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		result, err := operationalDays.GetOperationalDay(r.Context(), r.PathValue("operationalDayId"))
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, operationalDayResponse(result.Day))
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/operational-days/{operationalDayId}/summary",
		OperationID: "getOperationalDaySummary",
		Summary:     "Get operational day summary",
		Tags:        []string{"store-operations"},
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Operational day summary", Schema: operationalDaySummaryResponseSchema()},
			"400": {Description: "Invalid operational day summary query", Schema: httpapi.ProblemSchema()},
			"404": {Description: "Operational day was not found", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		summary, err := operationalDays.GetOperationalDaySummary(r.Context(), r.PathValue("operationalDayId"))
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, operationalDaySummaryResponse(summary))
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:          http.MethodGet,
		Path:            "/v1/operational-days/{operationalDayId}/receipts",
		OperationID:     "listOperationalDayReceipts",
		Summary:         "List receipts for operational day",
		Tags:            []string{"checkout", "store-operations"},
		QueryParameters: paginationQueryParams(),
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Operational day receipts", Schema: paginatedReceiptsResponseSchema()},
			"400": {Description: "Invalid operational day receipt query", Schema: httpapi.ProblemSchema()},
			"404": {Description: "Operational day was not found", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		operationalDayID := r.PathValue("operationalDayId")
		if _, err := operationalDays.GetOperationalDay(r.Context(), operationalDayID); err != nil {
			writeAppError(w, err)
			return
		}
		params := app.ParsePageParams(r.URL.Query().Get("limit"), r.URL.Query().Get("offset"))
		result, err := checkout.ListReceiptsByOperationalDay(r.Context(), operationalDayID, params)
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, PaginatedReceiptsResponse{
			Items:      receiptResponses(result.Items),
			TotalCount: result.TotalCount,
		})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:          http.MethodGet,
		Path:            "/v1/operational-days/{operationalDayId}/shifts",
		OperationID:     "listOperationalDayShifts",
		Summary:         "List shifts for operational day",
		Tags:            []string{"store-operations"},
		QueryParameters: paginationQueryParams(),
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Operational day shifts", Schema: paginatedShiftsResponseSchema()},
			"400": {Description: "Invalid operational day shift query", Schema: httpapi.ProblemSchema()},
			"404": {Description: "Operational day was not found", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		operationalDayID := r.PathValue("operationalDayId")
		if _, err := operationalDays.GetOperationalDay(r.Context(), operationalDayID); err != nil {
			writeAppError(w, err)
			return
		}
		params := app.ParsePageParams(r.URL.Query().Get("limit"), r.URL.Query().Get("offset"))
		result, err := shifts.ListShiftsByOperationalDay(r.Context(), operationalDayID, params)
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, PaginatedShiftsResponse{
			Items:      shiftResponses(result.Items),
			TotalCount: result.TotalCount,
		})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/stores/{storeId}/operational-days/current",
		OperationID: "getCurrentOperationalDay",
		Summary:     "Get current operational day",
		Tags:        []string{"store-operations"},
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Current operational day", Schema: operationalDayResponseSchema()},
			"404": {Description: "Open operational day was not found", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		result, err := operationalDays.GetCurrentOperationalDay(r.Context(), r.PathValue("storeId"))
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, operationalDayResponse(result.Day))
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodPost,
		Path:        "/v1/operational-days/{operationalDayId}/close-check",
		OperationID: "checkOperationalDayCloseReadiness",
		Summary:     "Check operational day close readiness",
		Tags:        []string{"store-operations"},
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Close readiness", Schema: operationalDayCloseReadinessResponseSchema()},
			"400": {Description: "Invalid operational day query", Schema: httpapi.ProblemSchema()},
			"404": {Description: "Operational day was not found", Schema: httpapi.ProblemSchema()},
			"409": {Description: "Operational day state conflict", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		result, err := operationalDays.CheckCloseReadiness(r.Context(), r.PathValue("operationalDayId"))
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, operationalDayCloseReadinessResponse(result))
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:              http.MethodPost,
		Path:                "/v1/operational-days/{operationalDayId}/close",
		OperationID:         "closeOperationalDay",
		Summary:             "Close operational day",
		Tags:                []string{"store-operations"},
		RequiresIdempotency: true,
		RequestBody: &httpapi.BodySpec{
			Description: "Operational day close command",
			Required:    true,
			Schema:      closeOperationalDayRequestSchema(),
		},
		Responses: map[string]httpapi.ResponseSpec{
			"202": {Description: "Operational day closed", Schema: operationalDayAcceptedResponseSchema()},
			"400": {Description: "Invalid operational day command", Schema: httpapi.ProblemSchema()},
			"404": {Description: "Operational day was not found", Schema: httpapi.ProblemSchema()},
			"409": {Description: "Operational day close blocked", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		if _, err := httpapi.RequireIdempotencyKey(r); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "idempotency_key_required", "Idempotency key is required", err.Error())
			return
		}
		var request CloseOperationalDayRequest
		if err := httpapi.DecodeJSON(r, &request); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_json", "Invalid JSON", err.Error())
			return
		}
		result, err := operationalDays.CloseOperationalDay(r.Context(), app.CloseOperationalDayCommand{
			IdempotencyKey:  r.Header.Get("Idempotency-Key"),
			DayID:           r.PathValue("operationalDayId"),
			ClosedByID:      request.ClosedByID,
			OverrideNoSales: request.OverrideNoSales,
			OverrideActorID: request.OverrideActorID,
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusAccepted, OperationalDayAcceptedResponse{
			OperationalDay: operationalDayResponse(result.Day),
		})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:              http.MethodPost,
		Path:                "/v1/shifts",
		OperationID:         "openShift",
		Summary:             "Open cashier shift",
		Tags:                []string{"store-operations"},
		RequiresIdempotency: true,
		RequestBody: &httpapi.BodySpec{
			Description: "Shift opening command",
			Required:    true,
			Schema:      openShiftRequestSchema(),
		},
		Responses: map[string]httpapi.ResponseSpec{
			"202": {Description: "Shift opened", Schema: shiftAcceptedResponseSchema()},
			"400": {Description: "Invalid shift command", Schema: httpapi.ProblemSchema()},
			"409": {Description: "Shift or idempotency conflict", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		if _, err := httpapi.RequireIdempotencyKey(r); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "idempotency_key_required", "Idempotency key is required", err.Error())
			return
		}
		var request OpenShiftRequest
		if err := httpapi.DecodeJSON(r, &request); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_json", "Invalid JSON", err.Error())
			return
		}
		result, err := shifts.OpenShift(r.Context(), app.OpenShiftCommand{
			IdempotencyKey:   r.Header.Get("Idempotency-Key"),
			StoreID:          request.StoreID,
			TerminalID:       request.TerminalID,
			CashierID:        request.CashierID,
			DrawerID:         request.DrawerID,
			OpeningCashMinor: request.OpeningCashMinor,
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusAccepted, ShiftAcceptedResponse{
			Shift: shiftResponse(result.Shift),
		})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/shifts/{shiftId}",
		OperationID: "getShift",
		Summary:     "Get shift",
		Tags:        []string{"store-operations"},
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Shift", Schema: shiftResponseSchema()},
			"404": {Description: "Shift was not found", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		result, err := shifts.GetShift(r.Context(), r.PathValue("shiftId"))
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, shiftResponse(result.Shift))
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/shifts/{shiftId}/receipts",
		OperationID: "listShiftReceipts",
		Summary:     "List receipts for shift",
		Tags:        []string{"checkout", "store-operations"},
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Shift receipts", Schema: receiptsResponseSchema()},
			"400": {Description: "Invalid shift receipt query", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		receipts, err := checkout.ListReceiptsByShift(r.Context(), r.PathValue("shiftId"))
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, ReceiptsResponse{
			Receipts: receiptResponses(receipts),
		})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:              http.MethodPost,
		Path:                "/v1/shifts/{shiftId}/close",
		OperationID:         "closeShift",
		Summary:             "Close cashier shift",
		Tags:                []string{"store-operations"},
		RequiresIdempotency: true,
		RequestBody: &httpapi.BodySpec{
			Description: "Shift close command",
			Required:    true,
			Schema:      closeShiftRequestSchema(),
		},
		Responses: map[string]httpapi.ResponseSpec{
			"202": {Description: "Shift closed", Schema: shiftAcceptedResponseSchema()},
			"400": {Description: "Invalid shift command", Schema: httpapi.ProblemSchema()},
			"404": {Description: "Shift was not found", Schema: httpapi.ProblemSchema()},
			"409": {Description: "Shift or idempotency conflict", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		if _, err := httpapi.RequireIdempotencyKey(r); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "idempotency_key_required", "Idempotency key is required", err.Error())
			return
		}
		var request CloseShiftRequest
		if err := httpapi.DecodeJSON(r, &request); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_json", "Invalid JSON", err.Error())
			return
		}
		result, err := shifts.CloseShift(r.Context(), app.CloseShiftCommand{
			IdempotencyKey:   r.Header.Get("Idempotency-Key"),
			ShiftID:          r.PathValue("shiftId"),
			ClosingCashMinor: request.ClosingCashMinor,
			SafeID:           request.SafeID,
			ActorID:          request.ActorID,
			ApprovedByID:     request.ApprovedByID,
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusAccepted, ShiftAcceptedResponse{
			Shift: shiftResponse(result.Shift),
		})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/stores/{storeId}/shifts/open",
		OperationID: "listOpenStoreShifts",
		Summary:     "List open store shifts",
		Tags:        []string{"store-operations"},
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Open shifts", Schema: shiftsResponseSchema()},
			"400": {Description: "Invalid shift query", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		result, err := shifts.ListOpenShiftsByStore(r.Context(), r.PathValue("storeId"))
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, ShiftsResponse{
			Shifts: shiftResponses(result),
		})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodPost,
		Path:        "/v1/receipts",
		OperationID: "openReceipt",
		Summary:     "Open a receipt",
		Tags:        []string{"checkout"},
		RequestBody: &httpapi.BodySpec{
			Description: "Receipt opening command",
			Required:    true,
			Schema:      openReceiptRequestSchema(),
		},
		RequiresIdempotency: true,
		Responses: map[string]httpapi.ResponseSpec{
			"202": {Description: "Receipt command accepted", Schema: receiptAcceptedResponseSchema()},
			"400": {Description: "Invalid command", Schema: httpapi.ProblemSchema()},
			"409": {Description: "Idempotency conflict", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		if _, err := httpapi.RequireIdempotencyKey(r); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "idempotency_key_required", "Idempotency key is required", err.Error())
			return
		}
		var request OpenReceiptRequest
		if err := httpapi.DecodeJSON(r, &request); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_json", "Invalid JSON", err.Error())
			return
		}
		result, err := checkout.OpenReceipt(r.Context(), app.OpenReceiptCommand{
			IdempotencyKey: r.Header.Get("Idempotency-Key"),
			StoreID:        request.StoreID,
			TerminalID:     request.TerminalID,
			CashierID:      request.CashierID,
			Channel:        request.Channel,
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusAccepted, ReceiptAcceptedResponse{
			Receipt: receiptResponse(result.Receipt),
		})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/receipts/{receiptId}",
		OperationID: "getReceipt",
		Summary:     "Get receipt",
		Tags:        []string{"checkout"},
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Receipt", Schema: receiptResponseSchema()},
			"404": {Description: "Receipt was not found", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		result, err := checkout.GetReceipt(r.Context(), r.PathValue("receiptId"))
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, receiptResponse(result.Receipt))
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:              http.MethodPost,
		Path:                "/v1/receipts/{receiptId}/cancel",
		OperationID:         "cancelReceipt",
		Summary:             "Cancel draft receipt",
		Tags:                []string{"checkout"},
		RequiresIdempotency: true,
		RequestBody: &httpapi.BodySpec{
			Description: "Receipt cancellation command",
			Required:    true,
			Schema:      cancelReceiptRequestSchema(),
		},
		Responses: map[string]httpapi.ResponseSpec{
			"202": {Description: "Receipt cancelled", Schema: receiptAcceptedResponseSchema()},
			"400": {Description: "Invalid cancellation command", Schema: httpapi.ProblemSchema()},
			"404": {Description: "Receipt was not found", Schema: httpapi.ProblemSchema()},
			"409": {Description: "Receipt or idempotency conflict", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		if _, err := httpapi.RequireIdempotencyKey(r); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "idempotency_key_required", "Idempotency key is required", err.Error())
			return
		}
		var request CancelReceiptRequest
		if err := httpapi.DecodeJSON(r, &request); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_json", "Invalid JSON", err.Error())
			return
		}
		result, err := checkout.CancelReceipt(r.Context(), app.CancelReceiptCommand{
			IdempotencyKey: r.Header.Get("Idempotency-Key"),
			ReceiptID:      r.PathValue("receiptId"),
			Reason:         request.Reason,
			ActorID:        request.ActorID,
			ApprovedByID:   request.ApprovedByID,
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusAccepted, ReceiptAcceptedResponse{
			Receipt: receiptResponse(result.Receipt),
		})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:              http.MethodPost,
		Path:                "/v1/stores/{storeId}/cash-movements",
		OperationID:         "createCashMovement",
		Summary:             "Create immutable cash movement",
		Tags:                []string{"cash-office"},
		RequiresIdempotency: true,
		RequestBody: &httpapi.BodySpec{
			Description: "Cash movement command",
			Required:    true,
			Schema:      createCashMovementRequestSchema(),
		},
		Responses: map[string]httpapi.ResponseSpec{
			"202": {Description: "Cash movement posted", Schema: cashMovementAcceptedResponseSchema()},
			"400": {Description: "Invalid cash movement command", Schema: httpapi.ProblemSchema()},
			"409": {Description: "Cash movement or idempotency conflict", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		if _, err := httpapi.RequireIdempotencyKey(r); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "idempotency_key_required", "Idempotency key is required", err.Error())
			return
		}
		var request CreateCashMovementRequest
		if err := httpapi.DecodeJSON(r, &request); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_json", "Invalid JSON", err.Error())
			return
		}
		result, err := cash.CreateCashMovement(r.Context(), app.CreateCashMovementCommand{
			IdempotencyKey:    r.Header.Get("Idempotency-Key"),
			StoreID:           r.PathValue("storeId"),
			Type:              request.Type,
			FromContainerID:   request.FromContainerID,
			FromContainerType: request.FromContainerType,
			ToContainerID:     request.ToContainerID,
			ToContainerType:   request.ToContainerType,
			AmountMinor:       request.AmountMinor,
			Currency:          request.Currency,
			Reason:            request.Reason,
			ActorID:           request.ActorID,
			ApprovedByID:      request.ApprovedByID,
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusAccepted, CashMovementAcceptedResponse{
			Movement: cashMovementResponse(result.Movement),
		})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:          http.MethodGet,
		Path:            "/v1/stores/{storeId}/cash-movements",
		OperationID:     "listCashMovements",
		Summary:         "List immutable cash movements",
		Tags:            []string{"cash-office"},
		QueryParameters: paginationQueryParams(),
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Cash movements", Schema: paginatedCashMovementsResponseSchema()},
			"400": {Description: "Invalid cash movement query", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		params := app.ParsePageParams(r.URL.Query().Get("limit"), r.URL.Query().Get("offset"))
		result, err := cash.ListCashMovements(r.Context(), r.PathValue("storeId"), params)
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, PaginatedCashMovementsResponse{
			Items:      cashMovementResponses(result.Items),
			TotalCount: result.TotalCount,
		})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/stores/{storeId}/cash-balances",
		OperationID: "listCashBalances",
		Summary:     "List derived cash container balances",
		Tags:        []string{"cash-office"},
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Cash balances", Schema: cashBalancesResponseSchema()},
			"400": {Description: "Invalid cash balance query", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		result, err := cash.ListCashBalances(r.Context(), r.PathValue("storeId"))
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, CashBalancesResponse{
			Balances: cashBalanceResponses(result),
		})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:              http.MethodPost,
		Path:                "/v1/stores/{storeId}/cash-recounts",
		OperationID:         "createCashRecount",
		Summary:             "Create cash recount",
		Tags:                []string{"cash-office"},
		RequiresIdempotency: true,
		RequestBody: &httpapi.BodySpec{
			Description: "Cash recount command",
			Required:    true,
			Schema:      createCashRecountRequestSchema(),
		},
		Responses: map[string]httpapi.ResponseSpec{
			"202": {Description: "Cash recount created", Schema: cashRecountAcceptedResponseSchema()},
			"400": {Description: "Invalid cash recount command", Schema: httpapi.ProblemSchema()},
			"409": {Description: "Cash recount or idempotency conflict", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		if _, err := httpapi.RequireIdempotencyKey(r); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "idempotency_key_required", "Idempotency key is required", err.Error())
			return
		}
		var request CreateCashRecountRequest
		if err := httpapi.DecodeJSON(r, &request); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_json", "Invalid JSON", err.Error())
			return
		}
		storeID := r.PathValue("storeId")
		businessDate := ""
		if currentDay, err := operationalDays.GetCurrentOperationalDay(r.Context(), storeID); err == nil {
			businessDate = currentDay.Day.BusinessDate
		}
		result, err := cash.CreateCashRecount(r.Context(), app.CreateCashRecountCommand{
			IdempotencyKey: r.Header.Get("Idempotency-Key"),
			StoreID:        storeID,
			BusinessDate:   businessDate,
			ContainerID:    request.ContainerID,
			ContainerType:  request.ContainerType,
			Currency:       request.Currency,
			CountedMinor:   request.CountedMinor,
			Reason:         request.Reason,
			ActorID:        request.ActorID,
			ApprovedByID:   request.ApprovedByID,
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusAccepted, CashRecountAcceptedResponse{
			Recount: cashRecountResponse(result.Recount),
		})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:          http.MethodGet,
		Path:            "/v1/stores/{storeId}/cash-recounts",
		OperationID:     "listCashRecounts",
		Summary:         "List cash recounts",
		Tags:            []string{"cash-office"},
		QueryParameters: paginationQueryParams(),
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Cash recounts", Schema: paginatedCashRecountsResponseSchema()},
			"400": {Description: "Invalid cash recount query", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		params := app.ParsePageParams(r.URL.Query().Get("limit"), r.URL.Query().Get("offset"))
		result, err := cash.ListCashRecounts(r.Context(), r.PathValue("storeId"), params)
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, PaginatedCashRecountsResponse{
			Items:      cashRecountResponses(result.Items),
			TotalCount: result.TotalCount,
		})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:              http.MethodPost,
		Path:                "/v1/stores/{storeId}/cash-recounts/{recountId}/resolve",
		OperationID:         "resolveCashRecount",
		Summary:             "Resolve cash recount discrepancy",
		Tags:                []string{"cash-office"},
		RequiresIdempotency: true,
		RequestBody: &httpapi.BodySpec{
			Description: "Cash recount resolution command",
			Required:    true,
			Schema:      resolveCashRecountRequestSchema(),
		},
		Responses: map[string]httpapi.ResponseSpec{
			"202": {Description: "Cash recount resolved", Schema: cashRecountAcceptedResponseSchema()},
			"400": {Description: "Invalid cash recount resolution command", Schema: httpapi.ProblemSchema()},
			"404": {Description: "Cash recount was not found", Schema: httpapi.ProblemSchema()},
			"409": {Description: "Cash recount resolution conflict", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		if _, err := httpapi.RequireIdempotencyKey(r); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "idempotency_key_required", "Idempotency key is required", err.Error())
			return
		}
		var request ResolveCashRecountRequest
		if err := httpapi.DecodeJSON(r, &request); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_json", "Invalid JSON", err.Error())
			return
		}
		result, err := cash.ResolveCashRecount(r.Context(), app.ResolveCashRecountCommand{
			IdempotencyKey: r.Header.Get("Idempotency-Key"),
			StoreID:        r.PathValue("storeId"),
			RecountID:      r.PathValue("recountId"),
			ResolutionNote: request.ResolutionNote,
			ActorID:        request.ActorID,
			ApprovedByID:   request.ApprovedByID,
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusAccepted, CashRecountAcceptedResponse{
			Recount: cashRecountResponse(result.Recount),
		})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:              http.MethodPost,
		Path:                "/v1/receipts/{receiptId}/fiscal-documents",
		OperationID:         "createReceiptFiscalDocument",
		Summary:             "Create fiscal document for receipt",
		Tags:                []string{"fiscalization"},
		RequiresIdempotency: true,
		RequestBody: &httpapi.BodySpec{
			Description: "Fiscalization command",
			Required:    true,
			Schema:      createFiscalDocumentRequestSchema(),
		},
		Responses: map[string]httpapi.ResponseSpec{
			"202": {Description: "Fiscal document command accepted", Schema: fiscalDocumentAcceptedResponseSchema()},
			"400": {Description: "Invalid fiscalization command", Schema: httpapi.ProblemSchema()},
			"404": {Description: "Receipt was not found", Schema: httpapi.ProblemSchema()},
			"409": {Description: "Fiscalization or idempotency conflict", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		if _, err := httpapi.RequireIdempotencyKey(r); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "idempotency_key_required", "Idempotency key is required", err.Error())
			return
		}
		var request CreateFiscalDocumentRequest
		if err := httpapi.DecodeJSON(r, &request); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_json", "Invalid JSON", err.Error())
			return
		}
		result, err := fiscalization.CreateFiscalDocument(r.Context(), app.CreateFiscalDocumentCommand{
			IdempotencyKey: r.Header.Get("Idempotency-Key"),
			ReceiptID:      r.PathValue("receiptId"),
			DeviceID:       request.DeviceID,
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusAccepted, FiscalDocumentAcceptedResponse{
			Document: fiscalDocumentResponse(result.Document),
		})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/receipts/{receiptId}/fiscal-documents",
		OperationID: "listReceiptFiscalDocuments",
		Summary:     "List receipt fiscal documents",
		Tags:        []string{"fiscalization"},
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Receipt fiscal documents", Schema: fiscalDocumentsResponseSchema()},
			"404": {Description: "Receipt was not found", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		result, err := fiscalization.ListReceiptFiscalDocuments(r.Context(), r.PathValue("receiptId"))
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, FiscalDocumentsResponse{
			Documents: fiscalDocumentResponses(result),
		})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:              http.MethodPost,
		Path:                "/v1/receipts/{receiptId}/payments",
		OperationID:         "createReceiptPayment",
		Summary:             "Create captured payment for receipt",
		Tags:                []string{"payments"},
		RequiresIdempotency: true,
		RequestBody: &httpapi.BodySpec{
			Description: "Captured payment command",
			Required:    true,
			Schema:      createPaymentRequestSchema(),
		},
		Responses: map[string]httpapi.ResponseSpec{
			"202": {Description: "Payment command accepted", Schema: paymentAcceptedResponseSchema()},
			"400": {Description: "Invalid payment command", Schema: httpapi.ProblemSchema()},
			"404": {Description: "Receipt was not found", Schema: httpapi.ProblemSchema()},
			"409": {Description: "Payment or idempotency conflict", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		if _, err := httpapi.RequireIdempotencyKey(r); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "idempotency_key_required", "Idempotency key is required", err.Error())
			return
		}
		var request CreatePaymentRequest
		if err := httpapi.DecodeJSON(r, &request); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_json", "Invalid JSON", err.Error())
			return
		}
		result, err := payments.CreatePayment(r.Context(), app.CreatePaymentCommand{
			IdempotencyKey:    r.Header.Get("Idempotency-Key"),
			ReceiptID:         r.PathValue("receiptId"),
			Method:            request.Method,
			AmountMinor:       request.AmountMinor,
			ProviderReference: request.ProviderReference,
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusAccepted, PaymentAcceptedResponse{
			Payment: paymentResponse(result.Payment),
		})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/receipts/{receiptId}/payments",
		OperationID: "listReceiptPayments",
		Summary:     "List receipt payments",
		Tags:        []string{"payments"},
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Receipt payments", Schema: paymentsResponseSchema()},
			"404": {Description: "Receipt was not found", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		result, err := payments.ListReceiptPayments(r.Context(), r.PathValue("receiptId"))
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, PaymentsResponse{
			Payments: paymentResponses(result),
		})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:              http.MethodPost,
		Path:                "/v1/receipts/{receiptId}/payments/{paymentId}/cancel",
		OperationID:         "cancelReceiptPayment",
		Summary:             "Cancel same-day card or cash payment for receipt",
		Tags:                []string{"payments"},
		RequiresIdempotency: true,
		RequestBody: &httpapi.BodySpec{
			Description: "Payment cancel command",
			Required:    false,
			Schema:      cancelPaymentRequestSchema(),
		},
		Responses: map[string]httpapi.ResponseSpec{
			"202": {Description: "Payment cancelled", Schema: paymentAcceptedResponseSchema()},
			"400": {Description: "Invalid payment cancel command", Schema: httpapi.ProblemSchema()},
			"404": {Description: "Receipt or payment was not found", Schema: httpapi.ProblemSchema()},
			"409": {Description: "Payment cannot be cancelled", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		if _, err := httpapi.RequireIdempotencyKey(r); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "idempotency_key_required", "Idempotency key is required", err.Error())
			return
		}
		var request CancelPaymentRequest
		if r.ContentLength > 0 {
			if err := httpapi.DecodeJSON(r, &request); err != nil {
				httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_json", "Invalid JSON", err.Error())
				return
			}
		}
		result, err := payments.CancelPayment(r.Context(), app.CancelPaymentCommand{
			IdempotencyKey: r.Header.Get("Idempotency-Key"),
			ReceiptID:      r.PathValue("receiptId"),
			PaymentID:      r.PathValue("paymentId"),
			ActorID:        request.ActorID,
			Reason:         request.Reason,
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusAccepted, PaymentAcceptedResponse{
			Payment: paymentResponse(result.Payment),
		})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:              http.MethodPost,
		Path:                "/v1/receipts/{receiptId}/payments/{paymentId}/refund",
		OperationID:         "refundReceiptPayment",
		Summary:             "Refund captured card or cash payment for receipt",
		Tags:                []string{"payments"},
		RequiresIdempotency: true,
		RequestBody: &httpapi.BodySpec{
			Description: "Payment refund command",
			Required:    false,
			Schema:      refundPaymentRequestSchema(),
		},
		Responses: map[string]httpapi.ResponseSpec{
			"202": {Description: "Payment refunded", Schema: paymentAcceptedResponseSchema()},
			"400": {Description: "Invalid payment refund command", Schema: httpapi.ProblemSchema()},
			"404": {Description: "Receipt or payment was not found", Schema: httpapi.ProblemSchema()},
			"409": {Description: "Payment cannot be refunded", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		if _, err := httpapi.RequireIdempotencyKey(r); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "idempotency_key_required", "Idempotency key is required", err.Error())
			return
		}
		var request RefundPaymentRequest
		if r.ContentLength > 0 {
			if err := httpapi.DecodeJSON(r, &request); err != nil {
				httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_json", "Invalid JSON", err.Error())
				return
			}
		}
		result, err := payments.RefundPayment(r.Context(), app.RefundPaymentCommand{
			IdempotencyKey: r.Header.Get("Idempotency-Key"),
			ReceiptID:      r.PathValue("receiptId"),
			PaymentID:      r.PathValue("paymentId"),
			ActorID:        request.ActorID,
			Reason:         request.Reason,
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusAccepted, PaymentAcceptedResponse{
			Payment: paymentResponse(result.Payment),
		})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:              http.MethodPost,
		Path:                "/v1/receipts/{receiptId}/lines",
		OperationID:         "addReceiptLine",
		Summary:             "Add line to receipt",
		Tags:                []string{"checkout"},
		RequiresIdempotency: true,
		RequestBody: &httpapi.BodySpec{
			Description: "Receipt line command",
			Required:    true,
			Schema:      addReceiptLineRequestSchema(),
		},
		Responses: map[string]httpapi.ResponseSpec{
			"202": {Description: "Receipt line command accepted", Schema: receiptAcceptedResponseSchema()},
			"400": {Description: "Invalid command", Schema: httpapi.ProblemSchema()},
			"409": {Description: "Idempotency conflict", Schema: httpapi.ProblemSchema()},
			"404": {Description: "Receipt was not found", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		if _, err := httpapi.RequireIdempotencyKey(r); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "idempotency_key_required", "Idempotency key is required", err.Error())
			return
		}
		var request AddReceiptLineRequest
		if err := httpapi.DecodeJSON(r, &request); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_json", "Invalid JSON", err.Error())
			return
		}
		result, err := checkout.AddReceiptLine(r.Context(), app.AddReceiptLineCommand{
			IdempotencyKey: r.Header.Get("Idempotency-Key"),
			ReceiptID:      r.PathValue("receiptId"),
			ProductID:      request.ProductID,
			Barcode:        request.Barcode,
			Name:           request.Name,
			Quantity:       request.Quantity,
			UnitPriceMinor: request.UnitPriceMinor,
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusAccepted, ReceiptAcceptedResponse{
			Receipt: receiptResponse(result.Receipt),
		})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:              http.MethodPost,
		Path:                "/v1/receipts/{receiptId}/scan",
		OperationID:         "scanReceiptLine",
		Summary:             "Scan product into receipt",
		Tags:                []string{"checkout"},
		RequiresIdempotency: true,
		RequestBody: &httpapi.BodySpec{
			Description: "Scanned barcode command",
			Required:    true,
			Schema:      scanReceiptLineRequestSchema(),
		},
		Responses: map[string]httpapi.ResponseSpec{
			"202": {Description: "Scanned line command accepted", Schema: receiptAcceptedResponseSchema()},
			"400": {Description: "Invalid command", Schema: httpapi.ProblemSchema()},
			"404": {Description: "Receipt or product was not found", Schema: httpapi.ProblemSchema()},
			"409": {Description: "Idempotency conflict", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		if _, err := httpapi.RequireIdempotencyKey(r); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "idempotency_key_required", "Idempotency key is required", err.Error())
			return
		}
		var request ScanReceiptLineRequest
		if err := httpapi.DecodeJSON(r, &request); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_json", "Invalid JSON", err.Error())
			return
		}
		result, err := checkout.ScanReceiptLine(r.Context(), app.ScanReceiptLineCommand{
			IdempotencyKey: r.Header.Get("Idempotency-Key"),
			ReceiptID:      r.PathValue("receiptId"),
			Barcode:        request.Barcode,
			Quantity:       request.Quantity,
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusAccepted, ReceiptAcceptedResponse{
			Receipt: receiptResponse(result.Receipt),
		})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/catalog/products/by-barcode/{barcode}",
		OperationID: "findProductByBarcode",
		Summary:     "Find product by barcode",
		Tags:        []string{"catalog"},
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Product", Schema: productResponseSchema()},
			"404": {Description: "Product was not found", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		result, err := catalog.FindProductByBarcode(r.Context(), r.PathValue("barcode"))
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, productResponse(result.Product))
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:          http.MethodGet,
		Path:            "/v1/stores/{storeId}/terminals",
		OperationID:     "listStoreTerminals",
		Summary:         "List terminals for a store",
		Tags:            []string{"terminals"},
		QueryParameters: paginationQueryParams(),
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Store terminals", Schema: paginatedTerminalsResponseSchema()},
			"400": {Description: "Invalid list query", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		params := app.ParsePageParams(r.URL.Query().Get("limit"), r.URL.Query().Get("offset"))
		result, err := terminals.ListStoreTerminals(r.Context(), r.PathValue("storeId"), params)
		if err != nil {
			writeAppError(w, err)
			return
		}
		items := make([]TerminalResponse, 0, len(result.Items))
		for _, terminal := range result.Items {
			items = append(items, terminalResponse(terminal))
		}
		httpapi.WriteJSON(w, http.StatusOK, PaginatedTerminalsResponse{
			Items:      items,
			TotalCount: result.TotalCount,
		})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:              http.MethodPost,
		Path:                "/v1/terminals/{terminalId}/heartbeat",
		OperationID:         "recordTerminalHeartbeat",
		Summary:             "Record terminal heartbeat",
		Tags:                []string{"terminals"},
		RequiresIdempotency: true,
		RequestBody: &httpapi.BodySpec{
			Description: "Terminal heartbeat command",
			Required:    true,
			Schema:      terminalHeartbeatRequestSchema(),
		},
		Responses: map[string]httpapi.ResponseSpec{
			"202": {Description: "Heartbeat accepted", Schema: heartbeatResponseSchema()},
			"400": {Description: "Invalid heartbeat", Schema: httpapi.ProblemSchema()},
			"409": {Description: "Idempotency conflict", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		if _, err := httpapi.RequireIdempotencyKey(r); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "idempotency_key_required", "Idempotency key is required", err.Error())
			return
		}
		var request TerminalHeartbeatRequest
		if err := httpapi.DecodeJSON(r, &request); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_json", "Invalid JSON", err.Error())
			return
		}
		result, err := terminals.RecordHeartbeat(r.Context(), app.RecordTerminalHeartbeatCommand{
			IdempotencyKey:  r.Header.Get("Idempotency-Key"),
			TerminalID:      r.PathValue("terminalId"),
			StoreID:         request.StoreID,
			Kind:            request.Kind,
			SoftwareVersion: request.SoftwareVersion,
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusAccepted, HeartbeatResponse{
			Terminal: terminalResponse(result.Terminal),
		})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/terminals/{terminalId}",
		OperationID: "getTerminal",
		Summary:     "Get terminal state",
		Tags:        []string{"terminals"},
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Terminal state", Schema: terminalResponseSchema()},
			"404": {Description: "Terminal was not found", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		result, err := terminals.GetTerminal(r.Context(), r.PathValue("terminalId"))
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, terminalResponse(result.Terminal))
	})
}

func writeAppError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, app.ErrReceiptNotFound):
		httpapi.WriteProblem(w, http.StatusNotFound, "receipt_not_found", "Receipt was not found", err.Error())
	case errors.Is(err, app.ErrTerminalNotFound):
		httpapi.WriteProblem(w, http.StatusNotFound, "terminal_not_found", "Terminal was not found", err.Error())
	case errors.Is(err, app.ErrProductNotFound):
		httpapi.WriteProblem(w, http.StatusNotFound, "product_not_found", "Product was not found", err.Error())
	case errors.Is(err, app.ErrPaymentNotFound):
		httpapi.WriteProblem(w, http.StatusNotFound, "payment_not_found", "Payment was not found", err.Error())
	case errors.Is(err, app.ErrFiscalDocumentNotFound):
		httpapi.WriteProblem(w, http.StatusNotFound, "fiscal_document_not_found", "Fiscal document was not found", err.Error())
	case errors.Is(err, app.ErrShiftNotFound):
		httpapi.WriteProblem(w, http.StatusNotFound, "shift_not_found", "Shift was not found", err.Error())
	case errors.Is(err, app.ErrCashRecountNotFound):
		httpapi.WriteProblem(w, http.StatusNotFound, "cash_recount_not_found", "Cash recount was not found", err.Error())
	case errors.Is(err, app.ErrOperationalDayNotFound):
		httpapi.WriteProblem(w, http.StatusNotFound, "operational_day_not_found", "Operational day was not found", err.Error())
	case errors.Is(err, app.ErrIdempotencyKeyRequired):
		httpapi.WriteProblem(w, http.StatusBadRequest, "idempotency_key_required", "Idempotency key is required", err.Error())
	case errors.Is(err, app.ErrIdempotencyKeyReused):
		httpapi.WriteProblem(w, http.StatusConflict, "idempotency_key_reused", "Idempotency key was reused", err.Error())
	case errors.Is(err, app.ErrPaymentAmountExceedsRemaining):
		httpapi.WriteProblem(w, http.StatusConflict, "payment_amount_exceeds_remaining", "Payment amount exceeds receipt remaining amount", err.Error())
	case errors.Is(err, app.ErrPaymentCannotBeCancelled), errors.Is(err, domain.ErrPaymentCannotBeCancelled):
		httpapi.WriteProblem(w, http.StatusConflict, "payment_cannot_be_cancelled", "Payment cannot be cancelled", err.Error())
	case errors.Is(err, app.ErrPaymentCancelSameDayRequired):
		httpapi.WriteProblem(w, http.StatusConflict, "payment_cancel_same_day_required", "Payment cancel is allowed only on the receipt business date", err.Error())
	case errors.Is(err, app.ErrPaymentCancelNotSupported):
		httpapi.WriteProblem(w, http.StatusConflict, "payment_cancel_not_supported", "Payment cancel is not supported for this method", err.Error())
	case errors.Is(err, app.ErrPaymentCannotBeRefunded), errors.Is(err, domain.ErrPaymentCannotBeRefunded):
		httpapi.WriteProblem(w, http.StatusConflict, "payment_cannot_be_refunded", "Payment cannot be refunded", err.Error())
	case errors.Is(err, app.ErrPaymentRefundNotSupported):
		httpapi.WriteProblem(w, http.StatusConflict, "payment_refund_not_supported", "Payment refund is not supported for this method", err.Error())
	case errors.Is(err, app.ErrPaymentUseCancelInstead):
		httpapi.WriteProblem(w, http.StatusConflict, "payment_use_cancel_instead", "Use payment cancel for same-day pre-fiscal card payments", err.Error())
	case errors.Is(err, app.ErrCashDrawerRequired):
		httpapi.WriteProblem(w, http.StatusConflict, "cash_drawer_required", "Cash drawer is required for cash payment", err.Error())
	case errors.Is(err, app.ErrReceiptCannotBeCancelled), errors.Is(err, domain.ErrReceiptCannotBeCancelled):
		httpapi.WriteProblem(w, http.StatusConflict, "receipt_cannot_be_cancelled", "Receipt cannot be cancelled", err.Error())
	case errors.Is(err, app.ErrReceiptNotFullyPaid):
		httpapi.WriteProblem(w, http.StatusConflict, "receipt_not_fully_paid", "Receipt is not fully paid", err.Error())
	case errors.Is(err, app.ErrReceiptAlreadyFiscalized):
		httpapi.WriteProblem(w, http.StatusConflict, "receipt_already_fiscalized", "Receipt is already fiscalized", err.Error())
	case errors.Is(err, app.ErrSeparationOfDutiesViolation):
		httpapi.WriteProblem(w, http.StatusConflict, "separation_of_duties_violation", "Actor cannot approve their own critical operation", err.Error())
	case errors.Is(err, app.ErrCashRecountApprovalRequired):
		httpapi.WriteProblem(w, http.StatusConflict, "cash_recount_approval_required", "Cash recount discrepancy requires approval", err.Error())
	case errors.Is(err, app.ErrCashRecountResolutionNotNeeded):
		httpapi.WriteProblem(w, http.StatusConflict, "cash_recount_resolution_not_needed", "Cash recount resolution is not needed", err.Error())
	case errors.Is(err, app.ErrCashRecountAlreadyResolved):
		httpapi.WriteProblem(w, http.StatusConflict, "cash_recount_already_resolved", "Cash recount is already resolved", err.Error())
	case errors.Is(err, app.ErrOpenShiftRequired):
		httpapi.WriteProblem(w, http.StatusConflict, "open_shift_required", "Open cashier shift is required", err.Error())
	case errors.Is(err, app.ErrOpenOperationalDayRequired):
		httpapi.WriteProblem(w, http.StatusConflict, "open_operational_day_required", "Open operational day is required", err.Error())
	case errors.Is(err, app.ErrShiftAlreadyOpenForTerminal):
		httpapi.WriteProblem(w, http.StatusConflict, "shift_already_open_for_terminal", "Shift is already open for terminal", err.Error())
	case errors.Is(err, app.ErrShiftAlreadyOpenForCashier):
		httpapi.WriteProblem(w, http.StatusConflict, "shift_already_open_for_cashier", "Shift is already open for cashier", err.Error())
	case errors.Is(err, app.ErrShiftAlreadyClosed):
		httpapi.WriteProblem(w, http.StatusConflict, "shift_already_closed", "Shift is already closed", err.Error())
	case errors.Is(err, app.ErrShiftCashCollectionRequired):
		httpapi.WriteProblem(w, http.StatusConflict, "shift_cash_collection_required", "Shift cash collection details are required", err.Error())
	case errors.Is(err, app.ErrShiftCloseBlocked):
		httpapi.WriteProblem(w, http.StatusConflict, "shift_close_blocked", "Shift close is blocked", err.Error())
	case errors.Is(err, app.ErrOperationalDayAlreadyOpen):
		httpapi.WriteProblem(w, http.StatusConflict, "operational_day_already_open", "Operational day is already open for store", err.Error())
	case errors.Is(err, app.ErrOperationalDayAlreadyClosed):
		httpapi.WriteProblem(w, http.StatusConflict, "operational_day_already_closed", "Operational day is already closed", err.Error())
	case errors.Is(err, app.ErrOperationalDayCloseBlocked):
		httpapi.WriteProblem(w, http.StatusConflict, "operational_day_close_blocked", "Operational day close is blocked", err.Error())
	case errors.Is(err, app.ErrInvalidCheckoutCommand), errors.Is(err, domain.ErrInvalidReceiptInput):
		httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_checkout_command", "Invalid checkout command", err.Error())
	case errors.Is(err, app.ErrInvalidTerminalCommand), errors.Is(err, domain.ErrInvalidTerminalInput):
		httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_terminal_command", "Invalid terminal command", err.Error())
	case errors.Is(err, app.ErrInvalidCatalogQuery), errors.Is(err, domain.ErrInvalidProductInput):
		httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_catalog_query", "Invalid catalog query", err.Error())
	case errors.Is(err, app.ErrInvalidPaymentCommand), errors.Is(err, domain.ErrInvalidPaymentInput):
		httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_payment_command", "Invalid payment command", err.Error())
	case errors.Is(err, app.ErrInvalidFiscalizationCommand), errors.Is(err, domain.ErrInvalidFiscalDocumentInput):
		httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_fiscalization_command", "Invalid fiscalization command", err.Error())
	case errors.Is(err, app.ErrInvalidCashMovementCommand), errors.Is(err, domain.ErrInvalidCashMovementInput):
		httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_cash_movement_command", "Invalid cash movement command", err.Error())
	case errors.Is(err, app.ErrInvalidCashRecountCommand), errors.Is(err, domain.ErrInvalidCashRecountInput):
		httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_cash_recount_command", "Invalid cash recount command", err.Error())
	case errors.Is(err, app.ErrInvalidShiftCommand), errors.Is(err, domain.ErrInvalidShiftInput):
		httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_shift_command", "Invalid shift command", err.Error())
	case errors.Is(err, app.ErrInvalidOperationalDayCommand), errors.Is(err, domain.ErrInvalidOperationalDayInput):
		httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_operational_day_command", "Invalid operational day command", err.Error())
	case errors.Is(err, app.ErrInvalidCredentials):
		httpapi.WriteProblem(w, http.StatusUnauthorized, "invalid_credentials", "Invalid credentials", err.Error())
	case errors.Is(err, app.ErrInvalidAuthCommand):
		httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_auth_command", "Invalid auth command", err.Error())
	case errors.Is(err, app.ErrSessionNotFound), errors.Is(err, app.ErrSessionExpired):
		httpapi.WriteProblem(w, http.StatusUnauthorized, "session_invalid", "Session is invalid or expired", err.Error())
	case errors.Is(err, app.ErrPermissionDenied):
		httpapi.WriteProblem(w, http.StatusForbidden, "permission_denied", "Permission denied", err.Error())
	case errors.Is(err, app.ErrInvalidReturnCommand), errors.Is(err, app.ErrReceiptNotReturnable):
		httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_return_command", "Invalid return command", err.Error())
	case errors.Is(err, app.ErrReturnNotFound):
		httpapi.WriteProblem(w, http.StatusNotFound, "return_not_found", "Return was not found", err.Error())
	case errors.Is(err, app.ErrReturnAlreadySettled), errors.Is(err, app.ErrReturnSettlementNotAllowed),
		errors.Is(err, app.ErrReturnSettlementRequiresFullReceiptReturn), errors.Is(err, app.ErrReturnSettlementPaymentMismatch):
		httpapi.WriteProblem(w, http.StatusConflict, "return_settlement_conflict", "Return settlement conflict", err.Error())
	case errors.Is(err, app.ErrInvalidDiscountCommand):
		httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_discount_command", "Invalid discount command", err.Error())
	case errors.Is(err, app.ErrInvalidMarkingCommand):
		httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_marking_command", "Invalid marking command", err.Error())
	default:
		httpapi.WriteProblem(w, http.StatusInternalServerError, "internal_error", "Internal error", err.Error())
	}
}

func demoProducts() []domain.Product {
	products := []domain.Product{
		{
			ID:             "demo-milk-1",
			Name:           "Demo Milk 1L",
			Barcodes:       []string{"4600000000000"},
			UnitPriceMinor: 19999,
			TaxCategoryID:  "vat_20",
		},
		{
			ID:             "demo-bread-1",
			Name:           "Demo Bread",
			Barcodes:       []string{"4600000000001"},
			UnitPriceMinor: 5999,
			TaxCategoryID:  "vat_10",
		},
	}
	for i, product := range products {
		normalized, err := domain.NewProduct(product)
		if err != nil {
			panic(err)
		}
		products[i] = normalized
	}
	return products
}

func operationalDayResponse(day domain.OperationalDay) OperationalDayResponse {
	return OperationalDayResponse{
		ID:           day.ID,
		StoreID:      day.StoreID,
		BusinessDate: day.BusinessDate,
		Status:       day.Status,
		OpenedByID:   day.OpenedByID,
		ClosedByID:   day.ClosedByID,
		OpenedAt:     day.OpenedAt,
		ClosedAt:     day.ClosedAt,
		UpdatedAt:    day.UpdatedAt,
	}
}

func operationalDayCloseReadinessResponse(readiness app.OperationalDayCloseReadiness) OperationalDayCloseReadinessResponse {
	return OperationalDayCloseReadinessResponse{
		OperationalDay: operationalDayResponse(readiness.Day),
		CanClose:       readiness.CanClose,
		Blockers:       operationalDayBlockers(readiness.Blockers),
	}
}

func operationalDaySummaryResponse(summary app.OperationalDaySummary) OperationalDaySummaryResponse {
	return OperationalDaySummaryResponse{
		OperationalDay: operationalDayResponse(summary.Day),
		CanClose:       summary.CanClose,
		Blockers:       operationalDayBlockers(summary.Blockers),
		Shifts: OperationalDayShiftSummary{
			TotalCount:  summary.Shifts.TotalCount,
			OpenCount:   summary.Shifts.OpenCount,
			ClosedCount: summary.Shifts.ClosedCount,
		},
		Cash: OperationalDayCashSummary{
			Balances:           cashBalanceResponses(summary.Cash.Balances),
			NonZeroDrawerCount: summary.Cash.NonZeroDrawerCount,
			Recounts: CashRecountSummary{
				TotalCount:               summary.Cash.Recounts.TotalCount,
				BalancedCount:            summary.Cash.Recounts.BalancedCount,
				DiscrepancyCount:         summary.Cash.Recounts.DiscrepancyCount,
				OpenDiscrepancyCount:     summary.Cash.Recounts.OpenDiscrepancyCount,
				ResolvedDiscrepancyCount: summary.Cash.Recounts.ResolvedDiscrepancyCount,
			},
		},
		Receipts: OperationalDayReceiptSummary{
			TotalCount:           summary.Receipts.TotalCount,
			DraftCount:           summary.Receipts.DraftCount,
			PaymentStartedCount:  summary.Receipts.PaymentStartedCount,
			PaidCount:            summary.Receipts.PaidCount,
			FiscalizedCount:      summary.Receipts.FiscalizedCount,
			CancelledCount:       summary.Receipts.CancelledCount,
			UnresolvedCount:      summary.Receipts.UnresolvedCount,
			FiscalizedSalesMinor: summary.Receipts.FiscalizedSalesMinor,
		},
		Payments: operationalDayPaymentSummaryResponse(summary.Payments),
		Fiscal: OperationalDayFiscalSummary{
			TotalCount:           summary.Fiscal.TotalCount,
			FiscalizedCount:      summary.Fiscal.FiscalizedCount,
			FiscalizedTotalMinor: summary.Fiscal.FiscalizedTotalMinor,
		},
	}
}

func operationalDayPaymentSummaryResponse(summary app.OperationalDayPaymentSummary) OperationalDayPaymentSummary {
	methods := make([]OperationalDayPaymentMethod, 0, len(summary.Methods))
	for _, method := range summary.Methods {
		methods = append(methods, OperationalDayPaymentMethod{
			Method:             method.Method,
			CapturedCount:      method.CapturedCount,
			CapturedTotalMinor: method.CapturedTotalMinor,
		})
	}
	return OperationalDayPaymentSummary{
		TotalCount:         summary.TotalCount,
		CapturedCount:      summary.CapturedCount,
		CapturedTotalMinor: summary.CapturedTotalMinor,
		Methods:            methods,
	}
}

func operationalDayBlockers(blockers []domain.OperationalDayBlocker) []OperationalDayBlocker {
	result := make([]OperationalDayBlocker, 0, len(blockers))
	for _, blocker := range blockers {
		result = append(result, OperationalDayBlocker{
			Code:        blocker.Code,
			Severity:    blocker.Severity,
			Message:     blocker.Message,
			ReferenceID: blocker.ReferenceID,
		})
	}
	return result
}

func terminalResponse(terminal domain.Terminal) TerminalResponse {
	return TerminalResponse{
		ID:              terminal.ID,
		StoreID:         terminal.StoreID,
		Kind:            terminal.Kind,
		Status:          terminal.Status,
		SoftwareVersion: terminal.SoftwareVersion,
		LastSeenAt:      terminal.LastSeenAt,
		UpdatedAt:       terminal.UpdatedAt,
	}
}

func productResponse(product domain.Product) ProductResponse {
	return ProductResponse{
		ID:             product.ID,
		Name:           product.Name,
		Barcodes:       append([]string(nil), product.Barcodes...),
		UnitPriceMinor: product.UnitPriceMinor,
		TaxCategoryID:  product.TaxCategoryID,
		Active:         product.Active,
	}
}

func paymentResponse(payment domain.Payment) PaymentResponse {
	return PaymentResponse{
		ID:                payment.ID,
		ReceiptID:         payment.ReceiptID,
		Method:            payment.Method,
		Status:            payment.Status,
		AmountMinor:       payment.AmountMinor,
		ProviderReference: payment.ProviderReference,
		CreatedAt:         payment.CreatedAt,
		UpdatedAt:         payment.UpdatedAt,
		CapturedAt:        payment.CapturedAt,
	}
}

func fiscalDocumentResponse(document domain.FiscalDocument) FiscalDocumentResponse {
	return FiscalDocumentResponse{
		ID:           document.ID,
		ReceiptID:    document.ReceiptID,
		Kind:         document.Kind,
		Status:       document.Status,
		AmountMinor:  document.AmountMinor,
		DeviceID:     document.DeviceID,
		FiscalSign:   document.FiscalSign,
		FiscalizedAt: document.FiscalizedAt,
		CreatedAt:    document.CreatedAt,
	}
}

func fiscalDocumentResponses(documents []domain.FiscalDocument) []FiscalDocumentResponse {
	result := make([]FiscalDocumentResponse, 0, len(documents))
	for _, document := range documents {
		result = append(result, fiscalDocumentResponse(document))
	}
	return result
}

func cashMovementResponse(movement domain.CashMovement) CashMovementResponse {
	return CashMovementResponse{
		ID:                movement.ID,
		StoreID:           movement.StoreID,
		Type:              movement.Type,
		FromContainerID:   movement.FromContainerID,
		FromContainerType: movement.FromContainerType,
		ToContainerID:     movement.ToContainerID,
		ToContainerType:   movement.ToContainerType,
		AmountMinor:       movement.AmountMinor,
		Currency:          movement.Currency,
		Reason:            movement.Reason,
		ActorID:           movement.ActorID,
		ApprovedByID:      movement.ApprovedByID,
		Status:            movement.Status,
		CreatedAt:         movement.CreatedAt,
	}
}

func cashMovementResponses(movements []domain.CashMovement) []CashMovementResponse {
	result := make([]CashMovementResponse, 0, len(movements))
	for _, movement := range movements {
		result = append(result, cashMovementResponse(movement))
	}
	return result
}

func cashBalanceResponse(balance domain.CashBalance) CashBalanceResponse {
	return CashBalanceResponse{
		StoreID:        balance.StoreID,
		ContainerID:    balance.ContainerID,
		ContainerType:  balance.ContainerType,
		Currency:       balance.Currency,
		BalanceMinor:   balance.BalanceMinor,
		LastMovementAt: balance.LastMovementAt,
	}
}

func cashBalanceResponses(balances []domain.CashBalance) []CashBalanceResponse {
	result := make([]CashBalanceResponse, 0, len(balances))
	for _, balance := range balances {
		result = append(result, cashBalanceResponse(balance))
	}
	return result
}

func cashRecountResponse(recount domain.CashRecount) CashRecountResponse {
	return CashRecountResponse{
		ID:               recount.ID,
		StoreID:          recount.StoreID,
		BusinessDate:     recount.BusinessDate,
		ContainerID:      recount.ContainerID,
		ContainerType:    recount.ContainerType,
		Currency:         recount.Currency,
		ExpectedMinor:    recount.ExpectedMinor,
		CountedMinor:     recount.CountedMinor,
		DiscrepancyMinor: recount.DiscrepancyMinor,
		Reason:           recount.Reason,
		ActorID:          recount.ActorID,
		ApprovedByID:     recount.ApprovedByID,
		Status:           recount.Status,
		ResolutionStatus: recount.ResolutionStatus,
		ResolutionNote:   recount.ResolutionNote,
		ResolvedByID:     recount.ResolvedByID,
		ResolvedAt:       recount.ResolvedAt,
		CreatedAt:        recount.CreatedAt,
	}
}

func cashRecountResponses(recounts []domain.CashRecount) []CashRecountResponse {
	result := make([]CashRecountResponse, 0, len(recounts))
	for _, recount := range recounts {
		result = append(result, cashRecountResponse(recount))
	}
	return result
}

func shiftResponse(shift domain.Shift) ShiftResponse {
	return ShiftResponse{
		ID:               shift.ID,
		StoreID:          shift.StoreID,
		OperationalDayID: shift.OperationalDayID,
		BusinessDate:     shift.BusinessDate,
		TerminalID:       shift.TerminalID,
		CashierID:        shift.CashierID,
		DrawerID:         shift.DrawerID,
		Status:           shift.Status,
		OpeningCashMinor: shift.OpeningCashMinor,
		ClosingCashMinor: shift.ClosingCashMinor,
		OpenedAt:         shift.OpenedAt,
		ClosedAt:         shift.ClosedAt,
		UpdatedAt:        shift.UpdatedAt,
	}
}

func shiftResponses(shifts []domain.Shift) []ShiftResponse {
	result := make([]ShiftResponse, 0, len(shifts))
	for _, shift := range shifts {
		result = append(result, shiftResponse(shift))
	}
	return result
}

func paymentResponses(payments []domain.Payment) []PaymentResponse {
	result := make([]PaymentResponse, 0, len(payments))
	for _, payment := range payments {
		result = append(result, paymentResponse(payment))
	}
	return result
}

func receiptResponses(receipts []domain.Receipt) []ReceiptResponse {
	result := make([]ReceiptResponse, 0, len(receipts))
	for _, receipt := range receipts {
		result = append(result, receiptResponse(receipt))
	}
	return result
}

func receiptResponse(receipt domain.Receipt) ReceiptResponse {
	lines := make([]ReceiptLineResponse, 0, len(receipt.Lines))
	for _, line := range receipt.Lines {
		lines = append(lines, ReceiptLineResponse{
			ID:                  line.ID,
			ProductID:           line.ProductID,
			Barcode:             line.Barcode,
			Name:                line.Name,
			Quantity:            line.Quantity,
			UnitPriceMinor:      line.UnitPriceMinor,
			DiscountMinor:       line.DiscountMinor,
			DiscountReason:      line.DiscountReason,
			DiscountAppliedByID: line.DiscountAppliedByID,
			TotalMinor:          line.TotalMinor,
			AddedAt:             line.AddedAt,
		})
	}

	return ReceiptResponse{
		ID:                 receipt.ID,
		StoreID:            receipt.StoreID,
		OperationalDayID:   receipt.OperationalDayID,
		BusinessDate:       receipt.BusinessDate,
		ShiftID:            receipt.ShiftID,
		TerminalID:         receipt.TerminalID,
		CashierID:          receipt.CashierID,
		DrawerID:           receipt.DrawerID,
		Channel:            receipt.Channel,
		Status:             receipt.Status,
		Lines:              lines,
		CancelReason:       receipt.CancelReason,
		CancelledByID:      receipt.CancelledByID,
		CancelApprovedByID: receipt.CancelApprovedByID,
		CancelledAt:        receipt.CancelledAt,
		TotalMinor:         receipt.TotalMinor(),
		CreatedAt:          receipt.CreatedAt,
		UpdatedAt:          receipt.UpdatedAt,
	}
}

func statusResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"storeId":      httpapi.StringSchema(),
		"mode":         httpapi.StringSchema(),
		"businessDate": httpapi.StringSchema(),
		"generatedAt":  httpapi.DateTimeSchema(),
	}, "storeId", "mode", "businessDate", "generatedAt")
}

func outboxStatusResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"pendingCount":    {"type": "integer"},
		"publishedCount":  {"type": "integer"},
		"brokerConnected": {"type": "boolean"},
	}, "pendingCount", "publishedCount", "brokerConnected")
}

func openOperationalDayRequestSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"storeId":      httpapi.StringSchema(),
		"businessDate": httpapi.StringSchema(),
		"openedById":   httpapi.StringSchema(),
	}, "storeId", "businessDate", "openedById")
}

func closeOperationalDayRequestSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"closedById":      httpapi.StringSchema(),
		"overrideNoSales": {"type": "boolean"},
		"overrideActorId": httpapi.StringSchema(),
	}, "closedById")
}

func operationalDayAcceptedResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"operationalDay": operationalDayResponseSchema(),
	}, "operationalDay")
}

func operationalDayCloseReadinessResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"operationalDay": operationalDayResponseSchema(),
		"canClose":       {"type": "boolean"},
		"blockers":       httpapi.ArraySchema(operationalDayBlockerSchema()),
	}, "operationalDay", "canClose", "blockers")
}

func operationalDaySummaryResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"operationalDay": operationalDayResponseSchema(),
		"canClose":       {"type": "boolean"},
		"blockers":       httpapi.ArraySchema(operationalDayBlockerSchema()),
		"shifts":         operationalDayShiftSummarySchema(),
		"cash":           operationalDayCashSummarySchema(),
		"receipts":       operationalDayReceiptSummarySchema(),
		"payments":       operationalDayPaymentSummarySchema(),
		"fiscal":         operationalDayFiscalSummarySchema(),
	}, "operationalDay", "canClose", "blockers", "shifts", "cash", "receipts", "payments", "fiscal")
}

func operationalDayShiftSummarySchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"totalCount":  {"type": "integer"},
		"openCount":   {"type": "integer"},
		"closedCount": {"type": "integer"},
	}, "totalCount", "openCount", "closedCount")
}

func operationalDayCashSummarySchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"balances":           httpapi.ArraySchema(cashBalanceResponseSchema()),
		"nonZeroDrawerCount": {"type": "integer"},
		"recounts":           cashRecountSummarySchema(),
	}, "balances", "nonZeroDrawerCount", "recounts")
}

func cashRecountSummarySchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"totalCount":               {"type": "integer"},
		"balancedCount":            {"type": "integer"},
		"discrepancyCount":         {"type": "integer"},
		"openDiscrepancyCount":     {"type": "integer"},
		"resolvedDiscrepancyCount": {"type": "integer"},
	}, "totalCount", "balancedCount", "discrepancyCount", "openDiscrepancyCount", "resolvedDiscrepancyCount")
}

func operationalDayReceiptSummarySchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"totalCount":           {"type": "integer"},
		"draftCount":           {"type": "integer"},
		"paymentStartedCount":  {"type": "integer"},
		"paidCount":            {"type": "integer"},
		"fiscalizedCount":      {"type": "integer"},
		"cancelledCount":       {"type": "integer"},
		"unresolvedCount":      {"type": "integer"},
		"fiscalizedSalesMinor": {"type": "integer"},
	}, "totalCount", "draftCount", "paymentStartedCount", "paidCount", "fiscalizedCount", "cancelledCount", "unresolvedCount", "fiscalizedSalesMinor")
}

func operationalDayPaymentSummarySchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"totalCount":         {"type": "integer"},
		"capturedCount":      {"type": "integer"},
		"capturedTotalMinor": {"type": "integer"},
		"methods":            httpapi.ArraySchema(operationalDayPaymentMethodSchema()),
	}, "totalCount", "capturedCount", "capturedTotalMinor", "methods")
}

func operationalDayPaymentMethodSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"method":             httpapi.StringSchema(),
		"capturedCount":      {"type": "integer"},
		"capturedTotalMinor": {"type": "integer"},
	}, "method", "capturedCount", "capturedTotalMinor")
}

func operationalDayFiscalSummarySchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"totalCount":           {"type": "integer"},
		"fiscalizedCount":      {"type": "integer"},
		"fiscalizedTotalMinor": {"type": "integer"},
	}, "totalCount", "fiscalizedCount", "fiscalizedTotalMinor")
}

func operationalDayResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"id":           httpapi.StringSchema(),
		"storeId":      httpapi.StringSchema(),
		"businessDate": httpapi.StringSchema(),
		"status":       httpapi.StringSchema(),
		"openedById":   httpapi.StringSchema(),
		"closedById":   httpapi.StringSchema(),
		"openedAt":     httpapi.DateTimeSchema(),
		"closedAt":     httpapi.DateTimeSchema(),
		"updatedAt":    httpapi.DateTimeSchema(),
	}, "id", "storeId", "businessDate", "status", "openedById", "openedAt", "updatedAt")
}

func operationalDayBlockerSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"code":        httpapi.StringSchema(),
		"severity":    httpapi.StringSchema(),
		"message":     httpapi.StringSchema(),
		"referenceId": httpapi.StringSchema(),
	}, "code", "severity", "message")
}

func receiptAcceptedResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"receipt": receiptResponseSchema(),
	}, "receipt")
}

func receiptsResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"receipts": httpapi.ArraySchema(receiptResponseSchema()),
	}, "receipts")
}

func heartbeatResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"terminal": terminalResponseSchema(),
	}, "terminal")
}

func terminalHeartbeatRequestSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"storeId":         httpapi.StringSchema(),
		"kind":            httpapi.StringSchema(),
		"softwareVersion": httpapi.StringSchema(),
	}, "storeId", "kind")
}

func openShiftRequestSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"storeId":          httpapi.StringSchema(),
		"terminalId":       httpapi.StringSchema(),
		"cashierId":        httpapi.StringSchema(),
		"drawerId":         httpapi.StringSchema(),
		"openingCashMinor": {"type": "integer", "minimum": 0},
	}, "storeId", "terminalId", "cashierId", "drawerId", "openingCashMinor")
}

func closeShiftRequestSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"closingCashMinor": {"type": "integer", "minimum": 0},
		"safeId":           httpapi.StringSchema(),
		"actorId":          httpapi.StringSchema(),
		"approvedById":     httpapi.StringSchema(),
	}, "closingCashMinor")
}

func openReceiptRequestSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"storeId":    httpapi.StringSchema(),
		"terminalId": httpapi.StringSchema(),
		"cashierId":  httpapi.StringSchema(),
		"channel":    httpapi.StringSchema(),
	}, "storeId", "terminalId", "cashierId")
}

func addReceiptLineRequestSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"productId":      httpapi.StringSchema(),
		"barcode":        httpapi.StringSchema(),
		"name":           httpapi.StringSchema(),
		"quantity":       {"type": "integer", "minimum": 1},
		"unitPriceMinor": {"type": "integer", "minimum": 0},
	}, "productId", "name", "quantity", "unitPriceMinor")
}

func scanReceiptLineRequestSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"barcode":  httpapi.StringSchema(),
		"quantity": {"type": "integer", "minimum": 1},
	}, "barcode", "quantity")
}

func cancelReceiptRequestSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"reason":       httpapi.StringSchema(),
		"actorId":      httpapi.StringSchema(),
		"approvedById": httpapi.StringSchema(),
	}, "reason", "actorId")
}

func createPaymentRequestSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"method":            httpapi.StringSchema(),
		"amountMinor":       {"type": "integer", "minimum": 1},
		"providerReference": httpapi.StringSchema(),
	}, "method", "amountMinor")
}

func cancelPaymentRequestSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"actorId": httpapi.StringSchema(),
		"reason":  httpapi.StringSchema(),
	})
}

func refundPaymentRequestSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"actorId": httpapi.StringSchema(),
		"reason":  httpapi.StringSchema(),
	})
}

func createFiscalDocumentRequestSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"deviceId": httpapi.StringSchema(),
	}, "deviceId")
}

func createCashMovementRequestSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"type":              httpapi.StringSchema(),
		"fromContainerId":   httpapi.StringSchema(),
		"fromContainerType": httpapi.StringSchema(),
		"toContainerId":     httpapi.StringSchema(),
		"toContainerType":   httpapi.StringSchema(),
		"amountMinor":       {"type": "integer", "minimum": 1},
		"currency":          httpapi.StringSchema(),
		"reason":            httpapi.StringSchema(),
		"actorId":           httpapi.StringSchema(),
		"approvedById":      httpapi.StringSchema(),
	}, "type", "fromContainerId", "fromContainerType", "toContainerId", "toContainerType", "amountMinor", "actorId")
}

func createCashRecountRequestSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"containerId":   httpapi.StringSchema(),
		"containerType": httpapi.StringSchema(),
		"currency":      httpapi.StringSchema(),
		"countedMinor":  {"type": "integer", "minimum": 0},
		"reason":        httpapi.StringSchema(),
		"actorId":       httpapi.StringSchema(),
		"approvedById":  httpapi.StringSchema(),
	}, "containerId", "containerType", "countedMinor", "actorId")
}

func resolveCashRecountRequestSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"resolutionNote": httpapi.StringSchema(),
		"actorId":        httpapi.StringSchema(),
		"approvedById":   httpapi.StringSchema(),
	}, "resolutionNote", "actorId", "approvedById")
}

func receiptResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"id":                 httpapi.StringSchema(),
		"storeId":            httpapi.StringSchema(),
		"operationalDayId":   httpapi.StringSchema(),
		"businessDate":       httpapi.StringSchema(),
		"shiftId":            httpapi.StringSchema(),
		"terminalId":         httpapi.StringSchema(),
		"cashierId":          httpapi.StringSchema(),
		"drawerId":           httpapi.StringSchema(),
		"channel":            httpapi.StringSchema(),
		"status":             httpapi.StringSchema(),
		"lines":              httpapi.ArraySchema(receiptLineResponseSchema()),
		"cancelReason":       httpapi.StringSchema(),
		"cancelledById":      httpapi.StringSchema(),
		"cancelApprovedById": httpapi.StringSchema(),
		"cancelledAt":        httpapi.DateTimeSchema(),
		"totalMinor":         {"type": "integer"},
		"createdAt":          httpapi.DateTimeSchema(),
		"updatedAt":          httpapi.DateTimeSchema(),
	}, "id", "storeId", "terminalId", "cashierId", "channel", "status", "lines", "totalMinor", "createdAt", "updatedAt")
}

func receiptLineResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"id":                  httpapi.StringSchema(),
		"productId":           httpapi.StringSchema(),
		"barcode":             httpapi.StringSchema(),
		"name":                httpapi.StringSchema(),
		"quantity":            {"type": "integer"},
		"unitPriceMinor":      {"type": "integer"},
		"discountMinor":       {"type": "integer"},
		"discountReason":      httpapi.StringSchema(),
		"discountAppliedById": httpapi.StringSchema(),
		"totalMinor":          {"type": "integer"},
		"addedAt":             httpapi.DateTimeSchema(),
	}, "id", "productId", "name", "quantity", "unitPriceMinor", "totalMinor", "addedAt")
}

func terminalResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"id":              httpapi.StringSchema(),
		"storeId":         httpapi.StringSchema(),
		"kind":            httpapi.StringSchema(),
		"status":          httpapi.StringSchema(),
		"softwareVersion": httpapi.StringSchema(),
		"lastSeenAt":      httpapi.DateTimeSchema(),
		"updatedAt":       httpapi.DateTimeSchema(),
	}, "id", "storeId", "kind", "status", "lastSeenAt", "updatedAt")
}

func paginatedTerminalsResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"items":      httpapi.ArraySchema(terminalResponseSchema()),
		"totalCount": {"type": "integer"},
	}, "items", "totalCount")
}

func shiftAcceptedResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"shift": shiftResponseSchema(),
	}, "shift")
}

func shiftsResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"shifts": httpapi.ArraySchema(shiftResponseSchema()),
	}, "shifts")
}

func shiftResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"id":               httpapi.StringSchema(),
		"storeId":          httpapi.StringSchema(),
		"operationalDayId": httpapi.StringSchema(),
		"businessDate":     httpapi.StringSchema(),
		"terminalId":       httpapi.StringSchema(),
		"cashierId":        httpapi.StringSchema(),
		"drawerId":         httpapi.StringSchema(),
		"status":           httpapi.StringSchema(),
		"openingCashMinor": {"type": "integer"},
		"closingCashMinor": {"type": "integer"},
		"openedAt":         httpapi.DateTimeSchema(),
		"closedAt":         httpapi.DateTimeSchema(),
		"updatedAt":        httpapi.DateTimeSchema(),
	}, "id", "storeId", "terminalId", "cashierId", "drawerId", "status", "openingCashMinor", "closingCashMinor", "openedAt", "updatedAt")
}

func productResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"id":             httpapi.StringSchema(),
		"name":           httpapi.StringSchema(),
		"barcodes":       httpapi.ArraySchema(httpapi.StringSchema()),
		"unitPriceMinor": {"type": "integer"},
		"taxCategoryId":  httpapi.StringSchema(),
		"active":         {"type": "boolean"},
	}, "id", "name", "barcodes", "unitPriceMinor", "active")
}

func paymentAcceptedResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"payment": paymentResponseSchema(),
	}, "payment")
}

func paymentsResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"payments": httpapi.ArraySchema(paymentResponseSchema()),
	}, "payments")
}

func paymentResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"id":                httpapi.StringSchema(),
		"receiptId":         httpapi.StringSchema(),
		"method":            httpapi.StringSchema(),
		"status":            httpapi.StringSchema(),
		"amountMinor":       {"type": "integer"},
		"providerReference": httpapi.StringSchema(),
		"createdAt":         httpapi.DateTimeSchema(),
		"updatedAt":         httpapi.DateTimeSchema(),
		"capturedAt":        httpapi.DateTimeSchema(),
	}, "id", "receiptId", "method", "status", "amountMinor", "createdAt", "updatedAt", "capturedAt")
}

func fiscalDocumentAcceptedResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"document": fiscalDocumentResponseSchema(),
	}, "document")
}

func fiscalDocumentsResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"documents": httpapi.ArraySchema(fiscalDocumentResponseSchema()),
	}, "documents")
}

func fiscalDocumentResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"id":           httpapi.StringSchema(),
		"receiptId":    httpapi.StringSchema(),
		"kind":         httpapi.StringSchema(),
		"status":       httpapi.StringSchema(),
		"amountMinor":  {"type": "integer"},
		"deviceId":     httpapi.StringSchema(),
		"fiscalSign":   httpapi.StringSchema(),
		"fiscalizedAt": httpapi.DateTimeSchema(),
		"createdAt":    httpapi.DateTimeSchema(),
	}, "id", "receiptId", "kind", "status", "amountMinor", "deviceId", "fiscalSign", "fiscalizedAt", "createdAt")
}

func cashMovementAcceptedResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"movement": cashMovementResponseSchema(),
	}, "movement")
}

func cashMovementsResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"movements": httpapi.ArraySchema(cashMovementResponseSchema()),
	}, "movements")
}

func cashBalancesResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"balances": httpapi.ArraySchema(cashBalanceResponseSchema()),
	}, "balances")
}

func cashRecountAcceptedResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"recount": cashRecountResponseSchema(),
	}, "recount")
}

func cashRecountsResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"recounts": httpapi.ArraySchema(cashRecountResponseSchema()),
	}, "recounts")
}

func cashMovementResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"id":                httpapi.StringSchema(),
		"storeId":           httpapi.StringSchema(),
		"type":              httpapi.StringSchema(),
		"fromContainerId":   httpapi.StringSchema(),
		"fromContainerType": httpapi.StringSchema(),
		"toContainerId":     httpapi.StringSchema(),
		"toContainerType":   httpapi.StringSchema(),
		"amountMinor":       {"type": "integer"},
		"currency":          httpapi.StringSchema(),
		"reason":            httpapi.StringSchema(),
		"actorId":           httpapi.StringSchema(),
		"approvedById":      httpapi.StringSchema(),
		"status":            httpapi.StringSchema(),
		"createdAt":         httpapi.DateTimeSchema(),
	}, "id", "storeId", "type", "fromContainerId", "fromContainerType", "toContainerId", "toContainerType", "amountMinor", "currency", "actorId", "status", "createdAt")
}

func cashBalanceResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"storeId":        httpapi.StringSchema(),
		"containerId":    httpapi.StringSchema(),
		"containerType":  httpapi.StringSchema(),
		"currency":       httpapi.StringSchema(),
		"balanceMinor":   {"type": "integer"},
		"lastMovementAt": httpapi.DateTimeSchema(),
	}, "storeId", "containerId", "containerType", "currency", "balanceMinor", "lastMovementAt")
}

func cashRecountResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"id":               httpapi.StringSchema(),
		"storeId":          httpapi.StringSchema(),
		"containerId":      httpapi.StringSchema(),
		"containerType":    httpapi.StringSchema(),
		"currency":         httpapi.StringSchema(),
		"expectedMinor":    {"type": "integer"},
		"countedMinor":     {"type": "integer"},
		"discrepancyMinor": {"type": "integer"},
		"reason":           httpapi.StringSchema(),
		"actorId":          httpapi.StringSchema(),
		"approvedById":     httpapi.StringSchema(),
		"status":           httpapi.StringSchema(),
		"resolutionStatus": httpapi.StringSchema(),
		"resolutionNote":   httpapi.StringSchema(),
		"resolvedById":     httpapi.StringSchema(),
		"resolvedAt":       httpapi.DateTimeSchema(),
		"createdAt":        httpapi.DateTimeSchema(),
	}, "id", "storeId", "containerId", "containerType", "currency", "expectedMinor", "countedMinor", "discrepancyMinor", "actorId", "status", "resolutionStatus", "createdAt")
}
