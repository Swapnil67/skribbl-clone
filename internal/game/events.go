package game

import (
	"encoding/json"
	"fmt"
	"time"
)

// * EventType defines the discriminator for WebSocket frames
type EventType string

const (
	EventDrawStroke  EventType = "DRAW_STROKE"
	EventClearCanvas EventType = "CLEAR_CANVAS"
	EventUndoStroke  EventType = "UNDO_STROKE"
	EventChatMessage EventType = "CHAT_MESSAGE"
	EventSystemAlert EventType = "SYSTEM_ALERT"
)

// * InboundEvent is the general envelope for incoming frames from clients
type InboundEvent struct {
	Type    EventType       `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// * OutboundEvent is the standard payload sent back to connected clients
type OutboundEvent struct {
	Type      EventType   `json:"type"`
	SenderID  string      `json:"sender_id,omitempty"`
	Timestamp int64       `json:"timestamp"`
	Payload   interface{} `json:"payload"`
}

// * StrokeMode defines whether the player is drawing or erasing
type StrokeMode string

const (
	ModePencil StrokeMode = "pencil"
	ModeEraser StrokeMode = "eraser"
)

// * DrawStrokePayload holds coordinate vectors and styling
type DrawStrokePayload struct {
	PrevX     float64    `json:"prev_x"`
	PrevY     float64    `json:"prev_y"`
	CurrX     float64    `json:"curr_x"`
	CurrY     float64    `json:"curr_y"`
	Color     string     `json:"color"`      // Hex string, e.g., "#FF0000"
	LineWidth float64    `json:"line_width"` // Thickness in pixels
	Mode      StrokeMode `json:"mode"`       // "pencil" or "eraser"
}

// * ClearCanvasPayload handles room-wide canvas wipeout
type ClearCanvasPayload struct {
	ClearedAt int64 `json:"cleared_at"`
}

// * UndoStrokePayload identifies which stroke or sequence to roll back
type UndoStrokePayload struct {
	StrokeID string `json:"stroke_id,omitempty"`
}

// * ChatMessagePayload handles chat text and guess submissions
type ChatMessagePayload struct {
	Text     string `json:"text"`
	IsGuess  bool   `json:"is_guess"`
	Username string `json:"username,omitempty"`
}

// * SystemAlertPayload broadcasts lobby/game notifications
type SystemAlertPayload struct {
	Message string `json:"message"`
	Level   string `json:"level"` // * "info", "warning", "success"
}

func ProcessIncomingMessage(senderID string, rawMessage []byte) (*OutboundEvent, error) {
	var env InboundEvent
	if err := json.Unmarshal(rawMessage, &env); err != nil {
		return nil, fmt.Errorf("invalid envelope JSON: %w", err)
	}

	outbound := &OutboundEvent{
		Type:      env.Type,
		SenderID:  senderID,
		Timestamp: time.Now().UnixMilli(),
	}

	switch env.Type {
	case EventDrawStroke:
		var stroke DrawStrokePayload
		if err := json.Unmarshal(env.Payload, &stroke); err != nil {
			return nil, fmt.Errorf("invalid draw stroke payload: %w", err)
		}
		// * Coordinate validation/sanitization
		outbound.Payload = stroke

	case EventClearCanvas:
		outbound.Payload = ClearCanvasPayload{
			ClearedAt: time.Now().UnixMilli(),
		}

	case EventUndoStroke:
		var undo UndoStrokePayload
		if err := json.Unmarshal(env.Payload, &undo); err != nil {
			return nil, fmt.Errorf("invalid undo stroke payload: %w", err)
		}
		outbound.Payload = undo

	case EventChatMessage:
		var chat ChatMessagePayload
		if err := json.Unmarshal(env.Payload, &chat); err != nil {
			return nil, fmt.Errorf("invalid chat payload: %w", err)
		}
		outbound.Payload = chat

	default:
		return nil, fmt.Errorf("unsupported event type: %s", env.Type)
	}

	return outbound, nil
}
