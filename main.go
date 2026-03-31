package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// ---------------- Models ----------------

type Message struct {
	Event   string `json:"event"`
	Message string `json:"message"`
}

type Geofence struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Coordinates [][]float64 `json:"coordinates"`
	Category    string      `json:"category"`
}

type Vehicle struct {
	ID            string `json:"id"`
	VehicleNumber string `json:"vehicle_number"`
	DriverName    string `json:"driver_name"`
	VehicleType   string `json:"vehicle_type"`
	Phone         string `json:"phone"`
	Status        string `json:"status"`
}

// ---------------- Storage (NO DB) ----------------

var geofences []Geofence
var vehicles = make(map[string]Vehicle)
var vehicleState = make(map[string]string)
var clients = make(map[*websocket.Conn]bool)

// ---------------- WebSocket ----------------

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// ---------------- MAIN ----------------

func main() {

	r := gin.Default()

	fmt.Println("🔥 SERVER STARTED (No DB Mode)")

	// ---------------- Home ----------------
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Server running 🚀",
		})
	})

	// ---------------- WebSocket ----------------
	r.GET("/ws/alerts", func(c *gin.Context) {

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			fmt.Println("❌ Upgrade error:", err)
			return
		}

		clients[conn] = true
		fmt.Println("🔵 WebSocket connected!")
	})

	// ---------------- Create Geofence ----------------
	r.POST("/geofences", func(c *gin.Context) {

		var g Geofence

		if err := c.BindJSON(&g); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		g.ID = "geo_" + g.Name
		geofences = append(geofences, g)

		c.JSON(http.StatusOK, g)
	})

	// ---------------- Create Vehicle ----------------
	r.POST("/vehicles", func(c *gin.Context) {

		var v Vehicle

		if err := c.BindJSON(&v); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		v.ID = "veh_" + v.VehicleNumber
		v.Status = "active"

		if _, exists := vehicles[v.ID]; exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Vehicle already exists"})
			return
		}

		vehicles[v.ID] = v

		fmt.Println("✅ Vehicle stored:", v)

		c.JSON(http.StatusOK, v)
	})

	// ---------------- Vehicle Location ----------------
	r.POST("/vehicles/location", func(c *gin.Context) {

		var body struct {
			VehicleID string  `json:"vehicle_id"`
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		}

		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		fmt.Println("📍 Location update:", body)

		c.JSON(http.StatusOK, gin.H{"status": "updated"})
	})

	// ---------------- EXIT VEHICLE ----------------
	r.POST("/vehicles/exit", func(c *gin.Context) {

		var body struct {
			VehicleID string `json:"vehicle_id"`
		}

		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		v, exists := vehicles[body.VehicleID]
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Vehicle not found"})
			return
		}

		v.Status = "exited"
		vehicles[body.VehicleID] = v
		vehicleState[body.VehicleID] = "outside"

		// WebSocket alert
		for client := range clients {
			client.WriteJSON(gin.H{
				"event":      "vehicle_exit",
				"vehicle_id": body.VehicleID,
				"message":    "Vehicle exited manually",
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Vehicle exit recorded",
		})
	})

	// ---------------- PORT FIX (IMPORTANT) ----------------
	port := os.Getenv("PORT")
if port == "" {
    port = "10000"
}

fmt.Println("RUnning on port:", port)
r.Run(":" + port)
}