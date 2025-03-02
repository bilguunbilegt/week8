package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sajari/regression"
)

// S3 Configuration
const (
	region     = "us-east-1"
	bucketName = "sage-bilguun"
	fileKey    = "train/energy_test_illinois.csv"
)

// Fetch Data from S3
func fetchCSVFromS3() ([][]string, error) {
	sess, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		return nil, fmt.Errorf("failed to start AWS session: %v", err)
	}

	svc := s3.New(sess)
	result, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileKey),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data from S3: %v", err)
	}

	// Read CSV content
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %v", err)
	}

	// Parse CSV
	reader := csv.NewReader(buf)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %v", err)
	}

	return records, nil
}

// Train the Model
func trainModel(data [][]string) ([]float64, error) {
	var r regression.Regression
	r.SetObserved("Energy_KWh")
	r.SetVar(0, "Population")
	r.SetVar(1, "Temperature")

	// Skip header and train model
	for i, row := range data {
		if i == 0 {
			continue // Skip header
		}
		var population, temperature, energy float64
		fmt.Sscanf(row[1], "%f", &population)   // Convert to float
		fmt.Sscanf(row[2], "%f", &temperature)  // Convert to float
		fmt.Sscanf(row[3], "%f", &energy)       // Convert to float
		r.Train(regression.DataPoint(energy, []float64{population, temperature}))
	}

	r.Run() // Train Model
	coefficients := r.GetCoeffs()

	return coefficients, nil
}

// Save Model to JSON
func saveModel(coefficients []float64) error {
	file, err := os.Create("model.json")
	if err != nil {
		return fmt.Errorf("failed to save model: %v", err)
	}
	defer file.Close()

	// Convert coefficients to JSON
	modelData, err := json.MarshalIndent(coefficients, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode model data: %v", err)
	}

	_, err = file.Write(modelData)
	return err
}

func main() {
	fmt.Println("Fetching data from S3...")
	data, err := fetchCSVFromS3()
	if err != nil {
		log.Fatalf("Error fetching CSV: %v", err)
	}

	fmt.Println("Training the model...")
	coefficients, err := trainModel(data)
	if err != nil {
		log.Fatalf("Error training model: %v", err)
	}

	fmt.Println("Saving trained model...")
	err = saveModel(coefficients)
	if err != nil {
		log.Fatalf("Error saving model: %v", err)
	}

	fmt.Println("✅ Model training complete. Model saved as model.json")
}
