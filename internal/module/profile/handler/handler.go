package handler

import (
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"sensor-service/config"
	"sensor-service/internal/module/profile/dto"
	profileUsecase "sensor-service/internal/module/profile/usecase"
	"sensor-service/internal/platform/helper"
	"sensor-service/internal/platform/httpengine/httpresponse"
)

type ProfileHandler struct {
	ProfileUseCase profileUsecase.IProfileUseCase
	clients        map[string]map[*websocket.Conn]bool // email -> connections
	mu             sync.RWMutex
	upgrader       websocket.Upgrader
}

// NewProfileHandler creates a new profile handler instance
// O(1) initialization
func NewProfileHandler(usecase profileUsecase.IProfileUseCase) *ProfileHandler {
	return &ProfileHandler{
		ProfileUseCase: usecase,
		clients:        make(map[string]map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for now
			},
		},
	}
}

// GetUserProfile handles GET /api/profile/:email
// O(1) handler delegation to usecase
func (h *ProfileHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimPrefix(r.URL.Path, "/api/profile/")
	if email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	// Extract caller role from JWT token
	callerRole := helper.GetRoleFromToken(r)
	if callerRole == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	profile, err := h.ProfileUseCase.GetUserProfile(email)
	if err != nil {
		log.Println("{GetUserProfile}{Error} : ", err)
		http.Error(w, err.Message, err.Code)
		return
	}

	helper.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"status_code": http.StatusOK,
		"message":     "Profile retrieved successfully",
		"data":        profile,
	})
}

// UpdateUserProfile handles PUT /api/profile/:email
// O(1) handler delegation to usecase
func (h *ProfileHandler) UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimPrefix(r.URL.Path, "/api/profile/")
	if email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	// Extract caller role from JWT token
	callerRole := helper.GetRoleFromToken(r)
	if callerRole == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var request config.HTTPRequest
	request.Body = make(map[string]interface{})

	if err := helper.DecodeJSON(r.Body, &request.Body); err != nil {
		log.Println("{UpdateUserProfile}{Decode}{Error} : ", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request body - only allow specific fields
	allowedFields := map[string]bool{"name": true, "avatar_url": true, "bio": true}
	for key := range request.Body {
		if !allowedFields[key] {
			http.Error(w, "Invalid field: "+key, http.StatusBadRequest)
			return
		}
	}

	profile, err := h.ProfileUseCase.UpdateUserProfile(email, request, callerRole)
	if err != nil {
		log.Println("{UpdateUserProfile}{Error} : ", err)
		http.Error(w, err.Message, err.Code)
		return
	}

	// Broadcast update to all connected clients for this email
	h.broadcastProfileUpdate(email, profile)

	helper.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"status_code": http.StatusOK,
		"message":     "Profile updated successfully",
		"data":        profile,
	})
}

// HandleProfileWebSocket handles WebSocket /ws/profile/:email
// O(1) connection setup, O(n) broadcast where n = connected clients
func (h *ProfileHandler) HandleProfileWebSocket(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimPrefix(r.URL.Path, "/ws/profile/")
	if email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	// Extract caller role from JWT token
	callerRole := helper.GetRoleFromToken(r)
	if callerRole == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("{HandleProfileWebSocket}{Upgrade}{Error} : ", err)
		return
	}
	defer conn.Close()

	// Register client
	h.mu.Lock()
	if h.clients[email] == nil {
		h.clients[email] = make(map[*websocket.Conn]bool)
	}
	h.clients[email][conn] = true
	h.mu.Unlock()

	log.Printf("{HandleProfileWebSocket}{Connected} email=%s", email)

	// Listen for messages from client
	for {
		var msg map[string]interface{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("{HandleProfileWebSocket}{Error} : %v", err)
			}
			break
		}

		// Handle incoming message (e.g., ping/keep-alive)
		if msgType, ok := msg["type"].(string); ok && msgType == "ping" {
			conn.WriteJSON(map[string]interface{}{
				"type":      "pong",
				"timestamp": time.Now().Unix(),
			})
		}
	}

	// Unregister client
	h.mu.Lock()
	delete(h.clients[email], conn)
	if len(h.clients[email]) == 0 {
		delete(h.clients, email)
	}
	h.mu.Unlock()

	log.Printf("{HandleProfileWebSocket}{Disconnected} email=%s", email)
}

// broadcastProfileUpdate sends profile update to all connected clients
// O(n) where n = number of connected clients for this email
func (h *ProfileHandler) broadcastProfileUpdate(email string, profile interface{}) {
	h.mu.RLock()
	clients, exists := h.clients[email]
	h.mu.RUnlock()

	if !exists || len(clients) == 0 {
		return
	}

	message := map[string]interface{}{
		"type":      "profile_update",
		"data":      profile,
		"timestamp": time.Now().Unix(),
	}

	h.mu.Lock()
	for client := range clients {
		go func(c *websocket.Conn) {
			if err := c.WriteJSON(message); err != nil {
				log.Printf("{broadcastProfileUpdate}{WriteError} : %v", err)
				h.mu.Lock()
				delete(clients, c)
				h.mu.Unlock()
			}
		}(client)
	}
	h.mu.Unlock()
}
