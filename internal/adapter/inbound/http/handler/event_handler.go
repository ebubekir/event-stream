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

// CreateEvent handles POST /events
// @Summary Create a new event
// @Description Creates a new event and persists it to the configured database
// @Tags events
// @Accept json
// @Produce json
// @Param event body dto.CreateEventRequest true "Event data"
// @Success 201 {object} dto.CreateEventResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
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

// CreateEventBatch handles POST /events/batch
// @Summary Create multiple events
// @Description Creates multiple events in a single batch operation
// @Tags events
// @Accept json
// @Produce json
// @Param events body dto.CreateEventBatchRequest true "Events data"
// @Success 201 {object} dto.CreateEventBatchResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
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

// RegisterRoutes registers event routes on the given router group
func (h *EventHandler) RegisterRoutes(rg *gin.RouterGroup) {
	events := rg.Group("/events")
	{
		events.POST("", h.CreateEvent)
		events.POST("/batch", h.CreateEventBatch)
	}
}
