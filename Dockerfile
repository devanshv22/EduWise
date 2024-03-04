
# Use an official Go runtime as a parent image


FROM --platform=linux/amd64 golang:1.20 AS backend

# Set the working directory in the container
WORKDIR /app/backend

# Copy the backend source code into the container
COPY backend .

# Build the backend executable
RUN go build -o main .

# Expose port 8080 to the outside world
EXPOSE 8080

# Define the command to run the backend executable
CMD ["./main"]

