package graphql

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/config"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/database"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/services"
)

// Resolver is the root GraphQL resolver
type Resolver struct {
	db         *database.Database
	subscriber *services.SubscriberService
	charging   *services.ChargingService
	es2Service *services.ES2Service
}

// NewResolver creates a new GraphQL resolver
func NewResolver(db *database.Database, cfg *config.Config) *Resolver {
	return &Resolver{
		db:         db,
		subscriber: services.NewSubscriberService(db, cfg),
		charging:   services.NewChargingService(db, cfg),
		es2Service: services.NewES2Service(&cfg.ES2),
	}
}

func parseID(id string) (uint, error) {
	var n int
	_, err := fmt.Sscanf(id, "Subscriber:%d", &n)
	if err != nil {
		return 0, fmt.Errorf("invalid subscriber ID format")
	}
	return uint(n), nil
}

func parseCursorOffset(cursor string) int {
	var offset int
	fmt.Sscanf(cursor, "%d", &offset)
	return offset
}

func strPtr(s string) *string { return new(s) }

func buildConnectionPageInfo(offset, count int, total int64) *PageInfo {
	var startCursor, endCursor *string
	if count > 0 {
		startCursor = strPtr(base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", offset+1))))
		endCursor = strPtr(base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", offset+count))))
	}
	return &PageInfo{
		HasNextPage:     int64(offset+count) < total,
		HasPreviousPage: offset > 0,
		StartCursor:     startCursor,
		EndCursor:       endCursor,
	}
}

func (r *Resolver) Subscriber(ctx context.Context, id string) (*models.Subscriber, error) {
	sid, err := parseID(id)
	if err != nil {
		return nil, err
	}
	return r.subscriber.GetSubscriber(ctx, sid)
}

func (r *Resolver) SubscriberByImsi(ctx context.Context, imsi string) (*models.Subscriber, error) {
	return r.subscriber.GetSubscriberByIMSI(ctx, models.IMSI(imsi))
}

func (r *Resolver) Subscribers(ctx context.Context, first *int, after *string, filter *SubscriberFilter, sort *SubscriberSort) (*SubscriberConnection, error) {
	limit := 20
	if first != nil {
		limit = *first
	}
	offset := 0
	if after != nil {
		cursor, err := base64.StdEncoding.DecodeString(*after)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor")
		}
		offset = parseCursorOffset(string(cursor))
	}

	resp, err := r.subscriber.ListSubscribers(ctx, &services.ListSubscribersRequest{
		Page:     offset/limit + 1,
		PageSize: limit,
	})
	if err != nil {
		return nil, err
	}

	edges := make([]*SubscriberEdge, len(resp.Subscribers))
	for i, sub := range resp.Subscribers {
		s := sub // avoid loop-variable capture
		edges[i] = &SubscriberEdge{
			Node:   &s,
			Cursor: base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", offset+i+1))),
		}
	}

	return &SubscriberConnection{
		Edges:      edges,
		PageInfo:   buildConnectionPageInfo(offset, len(resp.Subscribers), resp.Total),
		TotalCount: int(resp.Total),
	}, nil
}

func (r *Resolver) SubscriberAccount(ctx context.Context, imsi string) (*models.SubscriberAccount, error) {
	return r.subscriber.GetAccount(ctx, imsi)
}

func (r *Resolver) UsageEvents(ctx context.Context, first *int, after *string, imsi *string, filter *UsageEventFilter) (*UsageEventConnection, error) {
	limit := 50
	if first != nil {
		limit = *first
	}
	offset := 0
	if after != nil {
		cursor, err := base64.StdEncoding.DecodeString(*after)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor")
		}
		offset = parseCursorOffset(string(cursor))
	}

	where := buildUsageEventFilter(imsi, filter)
	events, total, err := r.charging.ListUsageEvents(ctx, limit, offset, where)
	if err != nil {
		return nil, err
	}

	edges := make([]*UsageEventEdge, len(events))
	for i, event := range events {
		edges[i] = &UsageEventEdge{
			Node:   event,
			Cursor: base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", offset+i+1))),
		}
	}

	return &UsageEventConnection{
		Edges:      edges,
		PageInfo:   buildConnectionPageInfo(offset, len(events), total),
		TotalCount: int(total),
	}, nil
}

func (r *Resolver) UsageStats(ctx context.Context, input UsageStatsInput) (*UsageStats, error) {
	stats, err := r.charging.GetUsageStats(ctx, input.IMSI, string(input.Period))
	if err != nil {
		return nil, err
	}
	return &UsageStats{
		DataUsage:  stats.DataUsage,
		VoiceUsage: stats.VoiceUsage,
		SmsUsage:   stats.SmsUsage,
		Cost:       stats.Cost,
		Period:     stats.Period,
		Trend: &UsageTrend{
			Direction:        TrendDirection(stats.Trend.Direction),
			Percentage:       stats.Trend.Percentage,
			PeriodOverPeriod: stats.Trend.PeriodOverPeriod,
		},
	}, nil
}

func (r *Resolver) RealTimeUsage(ctx context.Context, imsi string) (*RealTimeUsage, error) {
	usage, err := r.charging.GetRealTimeUsage(ctx, imsi)
	if err != nil {
		return nil, err
	}

	var cs *CurrentSession
	if usage.CurrentSession != nil {
		cs = &CurrentSession{
			SessionID: usage.CurrentSession.SessionID,
			StartTime: usage.CurrentSession.StartTime,
			DataUsed:  usage.CurrentSession.DataUsed,
			VoiceUsed: usage.CurrentSession.VoiceUsed,
			SmsUsed:   usage.CurrentSession.SmsUsed,
			Cost:      usage.CurrentSession.Cost,
		}
	}

	return &RealTimeUsage{
		CurrentSession: cs,
		TodayUsage: &TodayUsage{
			DataUsed:  usage.TodayUsage.DataUsed,
			VoiceUsed: usage.TodayUsage.VoiceUsed,
			SmsUsed:   usage.TodayUsage.SmsUsed,
			Cost:      usage.TodayUsage.Cost,
		},
	}, nil
}

func (r *Resolver) Invoices(ctx context.Context, first *int, after *string, imsi *string, status *models.InvoiceStatus) (*InvoiceConnection, error) {
	limit := 20
	if first != nil {
		limit = *first
	}
	offset := 0
	if after != nil {
		cursor, err := base64.StdEncoding.DecodeString(*after)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor")
		}
		offset = parseCursorOffset(string(cursor))
	}

	invoices, total, err := r.subscriber.ListInvoices(ctx, limit, offset, imsi, status)
	if err != nil {
		return nil, err
	}

	edges := make([]*InvoiceEdge, len(invoices))
	for i, inv := range invoices {
		edges[i] = &InvoiceEdge{
			Node:   inv,
			Cursor: base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", offset+i+1))),
		}
	}

	return &InvoiceConnection{
		Edges:      edges,
		PageInfo:   buildConnectionPageInfo(offset, len(invoices), total),
		TotalCount: int(total),
	}, nil
}

func (r *Resolver) Invoice(ctx context.Context, id string) (*models.Invoice, error) {
	return r.subscriber.GetInvoice(ctx, id)
}

func (r *Resolver) RatingPlans(ctx context.Context) ([]*models.RatingPlan, error) {
	return r.subscriber.ListRatingPlans(ctx)
}

func (r *Resolver) RatingPlan(ctx context.Context, planId string) (*models.RatingPlan, error) {
	return r.subscriber.GetRatingPlan(ctx, planId)
}

func (r *Resolver) Alerts(ctx context.Context, first *int, after *string, subscriberId *int, severity *models.AlertSeverity, resolved *bool) (*AlertConnection, error) {
	limit := 50
	if first != nil {
		limit = *first
	}
	offset := 0
	if after != nil {
		cursor, err := base64.StdEncoding.DecodeString(*after)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor")
		}
		offset = parseCursorOffset(string(cursor))
	}

	alerts, total, err := r.subscriber.ListAlerts(ctx, limit, offset, subscriberId, severity, resolved)
	if err != nil {
		return nil, err
	}

	edges := make([]*AlertEdge, len(alerts))
	for i, alert := range alerts {
		edges[i] = &AlertEdge{
			Node:   alert,
			Cursor: base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", offset+i+1))),
		}
	}

	return &AlertConnection{
		Edges:      edges,
		PageInfo:   buildConnectionPageInfo(offset, len(alerts), total),
		TotalCount: int(total),
	}, nil
}

func (r *Resolver) SystemStats(ctx context.Context) (*models.SystemStats, error) {
	return r.charging.GetSystemStats(ctx)
}

func (r *Resolver) HealthStatus(ctx context.Context) (*models.HealthStatus, error) {
	return r.charging.GetHealthStatus(ctx)
}

func (r *Resolver) SearchSubscribers(ctx context.Context, query string, limit int) ([]*models.Subscriber, error) {
	return r.subscriber.SearchSubscribers(ctx, query, limit)
}

func (r *Resolver) SearchUsageEvents(ctx context.Context, query string, limit int) ([]*models.UsageEvent, error) {
	return r.charging.SearchUsageEvents(ctx, query, limit)
}

// --- Mutation resolvers ---

func (r *Resolver) CreateSubscriber(ctx context.Context, input CreateSubscriberInput) (*models.Subscriber, error) {
	req := &services.CreateSubscriberRequest{
		MSISDN:    input.MSISDN,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		PlanID:    uint(input.PlanID),
	}
	if input.OrganizationID != nil {
		req.OrganizationID = *input.OrganizationID
	}
	if input.EUICCID != nil {
		req.EUICCID = *input.EUICCID
	}
	return r.subscriber.CreateSubscriber(ctx, req)
}

func (r *Resolver) UpdateSubscriber(ctx context.Context, id string, input UpdateSubscriberInput) (*models.Subscriber, error) {
	sid, err := parseID(id)
	if err != nil {
		return nil, err
	}
	var planID *uint
	if input.PlanID != nil {
		p := uint(*input.PlanID)
		planID = &p
	}
	return r.subscriber.UpdateSubscriber(ctx, sid, &services.UpdateSubscriberRequest{
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		PlanID:    planID,
	})
}

func (r *Resolver) DeleteSubscriber(ctx context.Context, id string) (bool, error) {
	sid, err := parseID(id)
	if err != nil {
		return false, err
	}
	return r.subscriber.DeleteSubscriber(ctx, sid)
}

func (r *Resolver) TopUpBalance(ctx context.Context, imsi string, input TopUpInput) (*models.SubscriberAccount, error) {
	return r.subscriber.TopUpBalance(ctx, imsi, &models.TopUpRequest{
		Amount:          input.Amount,
		PaymentMethodID: input.PaymentMethodID,
	})
}

func (r *Resolver) SuspendSubscriber(ctx context.Context, id string) (*models.Subscriber, error) {
	sid, err := parseID(id)
	if err != nil {
		return nil, err
	}
	if err := r.subscriber.SuspendSubscriber(ctx, sid); err != nil {
		return nil, err
	}
	return r.subscriber.GetSubscriber(ctx, sid)
}

func (r *Resolver) ActivateSubscriber(ctx context.Context, id string) (*models.Subscriber, error) {
	sid, err := parseID(id)
	if err != nil {
		return nil, err
	}
	return r.subscriber.ActivateSubscriber(ctx, sid)
}

func (r *Resolver) TerminateSubscriber(ctx context.Context, id string) (*models.Subscriber, error) {
	sid, err := parseID(id)
	if err != nil {
		return nil, err
	}
	if err := r.subscriber.TerminateSubscriber(ctx, sid); err != nil {
		return nil, err
	}
	return r.subscriber.GetSubscriber(ctx, sid)
}

func (r *Resolver) ProvisionESIM(ctx context.Context, imsi string) (*ESIMProvisionResult, error) {
	// Fetch subscriber to pass to ES2 service
	sub, err := r.subscriber.GetSubscriberByIMSI(ctx, models.IMSI(imsi))
	if err != nil {
		return nil, fmt.Errorf("subscriber not found for IMSI %s: %w", imsi, err)
	}
	result, err := r.es2Service.ProvisionProfile(ctx, sub)
	if err != nil {
		return nil, err
	}
	return &ESIMProvisionResult{
		ProfileID:      result.ProfileID,
		ActivationCode: result.Activation.ActivationCode,
	}, nil
}

func (r *Resolver) ActivateESIM(ctx context.Context, imsi string, profileId string) (bool, error) {
	err := r.es2Service.ActivateProfile(ctx, imsi, profileId)
	return err == nil, err
}

func (r *Resolver) DeactivateESIM(ctx context.Context, imsi string, profileId string) (bool, error) {
	err := r.es2Service.DeactivateProfile(ctx, imsi, profileId)
	return err == nil, err
}

func (r *Resolver) AddPaymentMethod(ctx context.Context, subscriberId int, input PaymentMethodInput) (*models.PaymentMethod, error) {
	isDefault := false
	if input.IsDefault != nil {
		isDefault = *input.IsDefault
	}
	return r.subscriber.AddPaymentMethod(ctx, subscriberId, &models.AddPaymentMethodRequest{
		Type:      models.PaymentMethodType(input.Type),
		Token:     input.Token,
		IsDefault: isDefault,
	})
}

func (r *Resolver) RemovePaymentMethod(ctx context.Context, paymentMethodId string) (bool, error) {
	return r.subscriber.RemovePaymentMethod(ctx, paymentMethodId)
}

func (r *Resolver) SetDefaultPaymentMethod(ctx context.Context, paymentMethodId string) (*models.PaymentMethod, error) {
	return r.subscriber.SetDefaultPaymentMethod(ctx, paymentMethodId)
}

func (r *Resolver) ResolveAlert(ctx context.Context, alertId string) (*models.Alert, error) {
	return r.subscriber.ResolveAlert(ctx, alertId)
}

func (r *Resolver) CreateAlert(ctx context.Context, input CreateAlertInput) (*models.Alert, error) {
	return r.subscriber.CreateAlert(ctx, &models.CreateAlertRequest{
		Type:         models.AlertType(input.Type),
		Severity:     models.AlertSeverity(input.Severity),
		Message:      input.Message,
		SubscriberID: input.SubscriberID,
	})
}

func (r *Resolver) TriggerSystemMaintenance(ctx context.Context) (bool, error) {
	return r.charging.TriggerMaintenance(ctx)
}

// --- Subscription resolvers ---

func (r *Resolver) SubscriberUpdates(ctx context.Context, subscriberId string) (<-chan *models.Subscriber, error) {
	return r.subscriber.SubscribeToSubscriberUpdates(ctx, subscriberId)
}

func (r *Resolver) UsageUpdates(ctx context.Context, imsi string) (<-chan *models.UsageEvent, error) {
	return r.charging.SubscribeToUsageUpdates(ctx, imsi)
}

func (r *Resolver) AlertUpdates(ctx context.Context, severity models.AlertSeverity) (<-chan *models.Alert, error) {
	return r.subscriber.SubscribeToAlertUpdates(ctx, severity)
}

func (r *Resolver) SystemStatsUpdates(ctx context.Context) (<-chan *models.SystemStats, error) {
	return r.charging.SubscribeToSystemStatsUpdates(ctx)
}

// --- Filter / sort helpers ---

func buildUsageEventFilter(imsi *string, filter *UsageEventFilter) string {
	conditions := []string{}
	if imsi != nil {
		conditions = append(conditions, fmt.Sprintf("imsi = '%s'", *imsi))
	}
	if filter != nil {
		if filter.UsageType != nil {
			conditions = append(conditions, fmt.Sprintf("usage_type = '%s'", *filter.UsageType))
		}
		if filter.TimestampGte != nil {
			conditions = append(conditions, fmt.Sprintf("timestamp >= '%s'", filter.TimestampGte.Format(time.RFC3339)))
		}
		if filter.TimestampLte != nil {
			conditions = append(conditions, fmt.Sprintf("timestamp <= '%s'", filter.TimestampLte.Format(time.RFC3339)))
		}
		if filter.CostMin != nil {
			conditions = append(conditions, fmt.Sprintf("cost >= %f", *filter.CostMin))
		}
		if filter.CostMax != nil {
			conditions = append(conditions, fmt.Sprintf("cost <= %f", *filter.CostMax))
		}
	}
	if len(conditions) > 0 {
		return "WHERE " + strings.Join(conditions, " AND ")
	}
	return ""
}

// --- ResolverRoot interface implementation ---

func (r *Resolver) Query() QueryResolver               { return r }
func (r *Resolver) Mutation() MutationResolver         { return r }
func (r *Resolver) Subscription() SubscriptionResolver { return r }

// ChargingSessions placeholder query resolver
func (r *Resolver) ChargingSessions(ctx context.Context, first *int, after *string, imsi *string, status *SessionStatus) (*ChargingSessionConnection, error) {
	return &ChargingSessionConnection{
		Edges:      []*ChargingSessionEdge{},
		PageInfo:   &PageInfo{},
		TotalCount: 0,
	}, nil
}

// ChargingSession placeholder query resolver
func (r *Resolver) ChargingSession(ctx context.Context, sessionId string) (*ChargingSession, error) {
	return nil, fmt.Errorf("charging session %s not found", sessionId)
}

// --- GraphQL handler setup ---

func SetupGraphQLHandler(router *gin.Engine, resolver *Resolver) {
	graphqlHandler := handler.NewDefaultServer(
		NewExecutableSchema(Config{Resolvers: resolver}),
	)
	router.POST("/graphql", gin.WrapH(graphqlHandler))
	if gin.Mode() == gin.DebugMode {
		router.GET("/graphql", gin.WrapH(playground.Handler("GraphQL Playground", "/graphql")))
	}
}
