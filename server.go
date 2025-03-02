package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// PredictionRequest represents input JSON structure
type PredictionRequest struct {
	Population  float64 `json:"population"`
	Temperature float64 `json:"temperature"`
}

// PredictionResponse represents output JSON structure
type PredictionResponse struct {
	EnergyKWh float64 `json:"predicted_energy_kwh"`
}

// Load Model Coefficients
func loadModel() ([]float64, error) {
	file, err := os.ReadFile("model.json")
	if err != nil {
		log.Printf("Error loading model file: %v", err)
		return nil, err
	}

	var coefficients []float64
	err = json.Unmarshal(file, &coefficients)
	if err != nil {
		log.Printf("Error parsing model file: %v", err)
		return nil, err
	}

	return coefficients, nil
}

// Compute Energy Consumption Prediction
func predictEnergy(population, temperature float64) float64 {
	coeff, err := loadModel()
	if err != nil {
		log.Println("Failed to load model, returning default prediction")
		return 1000
	}

	fmt.Println("Model Coefficients:", coeff)

	// Apply regression formula
	prediction := coeff[0] + (coeff[1] * population) + (coeff[2] * temperature)

	// Ensure non-negative predictions
	return math.Max(prediction, 0)
}

// API Endpoint for Predictions
func predictHandler(c *gin.Context) {
	var req PredictionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	energy := predictEnergy(req.Population, req.Temperature)
	c.JSON(http.StatusOK, PredictionResponse{EnergyKWh: energy})
}

// Main function to start API
func main() {
	router := gin.Default()
	router.POST("/predict", predictHandler)

	fmt.Println("✅ API Server is running on port 5000")
	router.Run("0.0.0.0:5000")
}
