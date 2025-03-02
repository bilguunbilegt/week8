package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

const sagemakerEndpoint = "https://runtime.sagemaker.us-east-1.amazonaws.com/endpoints/xgboost-energy-forecast-2024-03-02/invocations"

// Request structure
type PredictionRequest struct {
	Population  float64 `json:"population"`
	Temperature float64 `json:"temperature"`
}

// SageMaker Response Structure
type SageMakerResponse struct {
	Predictions [][]float64 `json:"predictions"`
}

// Predict Energy Consumption using SageMaker
func predictWithSageMaker(population, temperature float64) (float64, error) {
	payload := map[string]interface{}{
		"instances": []map[string]float64{
			{"population": population, "temperature": temperature},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest("POST", sagemakerEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("AWS_AUTH_TOKEN"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var sagemakerResponse SageMakerResponse
	err = json.Unmarshal(body, &sagemakerResponse)
	if err != nil {
		return 0, err
	}

	return sagemakerResponse.Predictions[0][0], nil
}

// API handler for predictions
func predictHandler(c *gin.Context) {
	var req PredictionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	energy, err := predictWithSageMaker(req.Population, req.Temperature)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate prediction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"predicted_energy_kwh": energy})
}

func main() {
	router := gin.Default()
	router.POST("/predict", predictHandler)

	fmt.Println("API Server is running on port 5000")
	router.Run("0.0.0.0:5000")
}
