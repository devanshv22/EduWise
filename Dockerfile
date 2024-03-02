# Use an official Go runtime as a parent image
FROM golang:1.17 AS backend

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

# Use an official Node.js runtime as a parent image
FROM node:20 AS frontend

# Set the working directory in the container
WORKDIR /app/frontend

# Copy the frontend source code into the container
COPY frontend .

# Install dependencies
RUN npm install

# Build the frontend
RUN npm run build

# Stage 2: Serve frontend with nginx server
FROM nginx:alpine

# Copy the built frontend files into the nginx directory
COPY --from=frontend /app/frontend/.next /usr/share/nginx/html

# Expose port 80 to the outside world
EXPOSE 80
