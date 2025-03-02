FROM golang:1.24

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o train_model train_model.go
RUN ./train_model
RUN go build -o energy-forecast server.go

EXPOSE 5000
CMD ["./energy-forecast"]
