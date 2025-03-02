# Use the latest Go 1.24 image
FROM golang:1.24

# Set working directory inside the container
WORKDIR /app

# Copy Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy all project files
COPY . .

# **Step 1: Build the training script (Triggers SageMaker Training)**
RUN go build -o train_model train_model.go

# **Step 2: Run the training script inside the container**
# This triggers model training in SageMaker and waits for deployment
RUN ./train_model

# **Step 3: Build the main API**
RUN go build -o energy-forecast server.go

# Expose API port
EXPOSE 5000

# Set environment variables for AWS credentials (Replace with IAM roles if needed)
ENV AWS_REGION="us-east-1"
ENV AWS_ACCESS_KEY_ID="AKIA4T4OBWQ2KNYUDK6Z"
ENV AWS_SECRET_ACCESS_KEY="0p+J06XvniRN2t2BNtQjuGWV7lvdQ9ho8F9PwJ4T"

# **Step 4: Start the API**
CMD ["./energy-forecast"]
