# Use an official Go runtime as a parent image
FROM --platform=linux/amd64 golang:1.20 AS backend

# Set the working directory for backend in the container
WORKDIR /app/backend

# Copy the backend source code into the container
COPY backend .

COPY .env /app/backend/.env

# Build the backend executable
RUN go build -o main .

# Expose port 8080 to the outside world for backend
EXPOSE 8080

# Define the command to run the backend executable
CMD ["./main"]

# Now, let's add steps for building and running the frontend
FROM --platform=linux/amd64 node:latest AS frontend

# Set the working directory for frontend in the container
WORKDIR /app/frontend

# Copy the frontend source code into the container
COPY frontend .

# Install dependencies for frontend
RUN yarn install

# Install additional dependencies for frontend
RUN yarn add @mui/material @emotion/react @emotion/styled

# Build the frontend
RUN yarn build

# Expose port 3000 to the outside world for frontend
EXPOSE 3000

# Define the command to run the frontend
CMD ["yarn", "start"]

# Since you want to run both, we can use a multi-stage build to combine both
# Start with the backend stage
FROM --platform=linux/amd64 backend AS combined

# Switch back to the root directory
WORKDIR /app

# Copy the built backend and frontend from their respective stages
COPY --from=backend /app/backend /app/backend
COPY --from=frontend /app/frontend /app/frontend
