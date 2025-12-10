package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ebubekir/event-stream/internal/adapter/inbound/http/dto"
	"github.com/ebubekir/event-stream/internal/application/event"
	"github.com/ebubekir/event-stream/pkg/response"
)

// EventHandler handles HTTP requests for events
type EventHandler struct {
	service *event.EventService
}

// NewEventHandler creates a new EventHandler
func NewEventHandler(service *event.EventService) *EventHandler {
	return &EventHandler{
		service: service,
	}
}

// CreateEvent
// @ID CreateEvent
// @Summary Create a new event
// @Description Creates a new event and persists it to the configured database
// @Tags events
// @Param event body dto.CreateEventRequest true "Event data"
// @Success 201 {object} dto.CreateEventResponse
// @Failure default {object} response.ApiError
// @Router /events [post]
func (h *EventHandler) CreateEvent(c *gin.Context) {
	var req dto.CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err)
		return
	}

	cmd := req.ToCommand()
	id, err := h.service.CreateEvent(c.Request.Context(), cmd)
	if err != nil {
		response.SystemError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.CreateEventResponse{ID: id})
}

// CreateEventBatch
// @ID CreateEventBatch
// @Summary Create multiple events
// @Description Creates multiple events in a single batch operation
// @Tags events
// @Param events body dto.CreateEventBatchRequest true "Events data"
// @Success 201 {object} dto.CreateEventBatchResponse
// @Failure default {object} response.ApiError
// @Router /events/batch [post]
func (h *EventHandler) CreateEventBatch(c *gin.Context) {
	var req dto.CreateEventBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err)
		return
	}

	cmds := make([]*event.CreateEventCommand, len(req.Events))
	for i, eventReq := range req.Events {
		cmds[i] = eventReq.ToCommand()
	}

	ids, err := h.service.CreateEvents(c.Request.Context(), cmds)
	if err != nil {
		response.SystemError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.CreateEventBatchResponse{IDs: ids})
}

// GetMetrics
// @ID GetMetrics
// @Summary Get event metrics
// @Description Retrieves aggregated metrics for events with optional grouping
// @Tags events
// @Param event_name query string true "Event name to filter by"
// @Param from query string false "Start timestamp (RFC3339 format)"
// @Param to query string false "End timestamp (RFC3339 format)"
// @Param group_by query string false "Aggregation type: channel, daily, hourly"
// @Success 200 {object} dto.GetMetricsResponse
// @Failure default {object} response.ApiError
// @Router /events/metrics [get]
func (h *EventHandler) GetMetrics(c *gin.Context) {
	var req dto.GetMetricsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err)
		return
	}

	query, err := req.ToQuery()
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	result, err := h.service.GetMetrics(c.Request.Context(), query)
	if err != nil {
		response.SystemError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.FromMetricsResultDTO(result))
}

// RegisterRoutes registers event routes on the given router group
func (h *EventHandler) RegisterRoutes(rg *gin.RouterGroup) {
	events := rg.Group("/events")
	{
		events.POST("", h.CreateEvent)
		events.POST("/batch", h.CreateEventBatch)
		events.GET("/metrics", h.GetMetrics)
	}
}
