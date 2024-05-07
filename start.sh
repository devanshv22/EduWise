#!/bin/bash

# Navigate to backend directory
cd backend

# Check if .env file exists, if not, create one
if [ ! -f .env ]; then
    touch .env
fi

# Append environment variables to .env file
echo "MONGO_URI=$MONGO_URI" > .env
echo "SMTP_USERNAME=$SMTP_USERNAME" >> .env
echo "SMTP_PASSWORD=$SMTP_PASSWORD" >> .env

# Build the frontend
cd ../frontend
npm install
npm run build

# Navigate back to the main directory
cd ..

# Build the Go backend
cd backend
go build -o app
cd ..

# Run the backend
./backend/app -port 8080 &

# Run the frontend on port 3000
cd frontend
npm start -- --port 3000 &

# Sleep to keep the script running
sleep infinity