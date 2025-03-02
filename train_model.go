package main

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sagemaker"
)

const (
	bucketName    = "sage-bilguun"
	fileKey       = "train/energy_test_illinois.csv"
	roleArn       = "arn:aws:iam::867344430132:role/service-role/AmazonSageMaker-ExecutionRole-20250302T162596"
	region        = "us-east-1"
	modelOutputS3 = "s3://sage-bilguun/model/"
	trainImageUri = "683313688378.dkr.ecr.us-east-1.amazonaws.com/xgboost:latest"
	instanceType  = "ml.m5.large"
	jobName       = "energy-forecast-training-job"
	endpointName  = "energy-forecast-endpoint"
)

// Function to Start a SageMaker Training Job
func startTrainingJob(sess *session.Session) error {
	svc := sagemaker.New(sess)

	inputDataConfig := []*sagemaker.Channel{
		{
			ChannelName: aws.String("train"),
			DataSource: &sagemaker.DataSource{
				S3DataSource: &sagemaker.S3DataSource{
					S3DataType: aws.String("S3Prefix"),
					S3Uri:      aws.String("s3://" + bucketName + "/train/"),
				},
			},
			ContentType: aws.String("csv"),
		},
	}

	trainingParams := &sagemaker.CreateTrainingJobInput{
		TrainingJobName: aws.String(jobName),
		AlgorithmSpecification: &sagemaker.AlgorithmSpecification{
			TrainingImage:     aws.String(trainImageUri),
			TrainingInputMode: aws.String("File"),
		},
		RoleArn:         aws.String(roleArn),
		InputDataConfig: inputDataConfig,
		OutputDataConfig: &sagemaker.OutputDataConfig{
			S3OutputPath: aws.String(modelOutputS3),
		},
		ResourceConfig: &sagemaker.ResourceConfig{
			InstanceType:   aws.String(instanceType),
			InstanceCount:  aws.Int64(1),
			VolumeSizeInGB: aws.Int64(10),
		},
		StoppingCondition: &sagemaker.StoppingCondition{
			MaxRuntimeInSeconds: aws.Int64(3600),
		},
	}

	_, err := svc.CreateTrainingJob(trainingParams)
	if err != nil {
		return fmt.Errorf("failed to create training job: %v", err)
	}

	fmt.Println("Training job started successfully:", jobName)
	return nil
}

// Function to Deploy the Model as an Endpoint
func deployModel(sess *session.Session) error {
	svc := sagemaker.New(sess)

	modelParams := &sagemaker.CreateModelInput{
		ModelName: aws.String(endpointName),
		PrimaryContainer: &sagemaker.ContainerDefinition{
			Image:        aws.String(trainImageUri),
			ModelDataUrl: aws.String(modelOutputS3 + jobName + "/output/model.tar.gz"),
		},
		ExecutionRoleArn: aws.String(roleArn),
	}

	_, err := svc.CreateModel(modelParams)
	if err != nil {
		return fmt.Errorf("failed to create model: %v", err)
	}

	endpointConfigParams := &sagemaker.CreateEndpointConfigInput{
		EndpointConfigName: aws.String(endpointName),
		ProductionVariants: []*sagemaker.ProductionVariant{
			{
				VariantName:          aws.String("AllTraffic"),
				ModelName:            aws.String(endpointName),
				InstanceType:         aws.String(instanceType),
				InitialInstanceCount: aws.Int64(1),
			},
		},
	}

	_, err = svc.CreateEndpointConfig(endpointConfigParams)
	if err != nil {
		return fmt.Errorf("failed to create endpoint config: %v", err)
	}

	endpointParams := &sagemaker.CreateEndpointInput{
		EndpointName:       aws.String(endpointName),
		EndpointConfigName: aws.String(endpointName),
	}

	_, err = svc.CreateEndpoint(endpointParams)
	if err != nil {
		return fmt.Errorf("failed to deploy model: %v", err)
	}

	fmt.Println("SageMaker Model Deployed at:", endpointName)
	return nil
}

func main() {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		log.Fatalf("Failed to start AWS session: %v", err)
	}

	fmt.Println("Starting SageMaker training job...")
	err = startTrainingJob(sess)
	if err != nil {
		log.Fatalf("Error in training job: %v", err)
	}

	fmt.Println("Deploying trained model as endpoint...")
	err = deployModel(sess)
	if err != nil {
		log.Fatalf("Error deploying model: %v", err)
	}
}
