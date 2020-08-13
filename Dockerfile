FROM golang:latest

WORKDIR /app
# Copy go mod and sum files
COPY go.mod go.sum ./
# Download all dependencies.
RUN go mod download

COPY . .

RUN  go build -race -ldflags "-extldflags '-static'" -o main

EXPOSE 8080

CMD ["./main"]
