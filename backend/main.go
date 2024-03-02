package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"net/smtp"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var (
	collection        *mongo.Collection
	facultyCollection *mongo.Collection
	courseCollection  *mongo.Collection
	registeredUsers   *mongo.Collection
	ctx               = context.TODO()
	jwtKey            = []byte("3J&59#sM%5D+^!Y$BXu@2pPw@sn#ZjF")
)

// Claims structure for JWT token
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

type UserRegistration struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	IsVerified bool   `json:"isVerified"`
	OTP        string `json:"otp"`
	// Add any additional fields for registration such as email, name, etc.
}

type File struct {
	ID          string `json:"id" bson:"_id"`
	CourseName  string `json:"courseName" bson:"courseName"`
	Batch       string `json:"batch" bson:"batch"`
	Instructor  string `json:"instructor" bson:"instructor"`
	Type        string `json:"type" bson:"type"`
	Remark      string `json:"remark" bson:"remark"`
	FileContent []byte `json:"fileContent" bson:"fileContent"`
	Link        string `json:"link" bson:"link"`
}

type Faculty struct {
	Name string `json:"name" bson:"name"`
}

type Course struct {
	Name string `json:"name" bson:"name"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		log.Fatal(err)
	}
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)
	db := client.Database("CoursesInfo")
	collection = db.Collection("details")

	facultyDB := client.Database("FacultyInfo")
	facultyCollection = facultyDB.Collection("details")

	courseDB := client.Database("ListofCourse")
	courseCollection = courseDB.Collection("details")

	registerDB := client.Database("Userdata")
	registeredUsers = registerDB.Collection("registered_users")

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, username") // Include 'username' header
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	r.POST("/api/upload", uploadFile)
	r.GET("/api/fetch", fetchFiles)
	r.GET("/api/faculty", fetchFaculty)
	r.POST("/api/faculty", uploadFaculty)
	r.POST("/api/courses", uploadCourse)
	r.GET("/api/courses", fetchCourses)
	r.POST("/api/login", login)
	r.POST("/api/register", register)
	r.POST("/api/verify", verifyOTP)
	r.GET("/api/isLoggedIn", isLoggedIn)
	r.GET("/api/main", isAuthenticated, isLoggedIn, mainPageHandler)
	r.GET("/api/download/:fileID", downloadFile)

	if err := r.Run("0.0.0.0:8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}

func generateFileLink(fileID string) string {
	// Assuming your frontend is hosted at http://localhost:3000
	return fmt.Sprintf("http://localhost:8080/api/download/%s?download=true", fileID)
}

func uploadFile(c *gin.Context) {
	// Extract the username from the request headers
	username := c.GetHeader("username")

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Open uploaded file
	fileContent, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file"})
		return
	}
	defer fileContent.Close()

	// Read file content
	content, err := ioutil.ReadAll(fileContent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file content"})
		return
	}

	// Generate a unique identifier for the file
	fileID := primitive.NewObjectID().Hex()

	courseName := c.PostForm("courseName")
	batch := c.PostForm("batch")
	instructor := c.PostForm("instructor")
	fileType := c.PostForm("type")
	remark := c.PostForm("remark")

	// Store file content and other details in the database along with the username
	_, err = collection.InsertOne(ctx, bson.M{
		"_id":         fileID, // Store the unique identifier
		"username":    username,
		"courseName":  courseName,
		"batch":       batch,
		"instructor":  instructor,
		"type":        fileType,
		"remark":      remark,
		"fileContent": content,                  // Store file content
		"link":        generateFileLink(fileID), // Store the generated link
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store data in database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Upload successful"})
}

// Function to extract the username from the JWT token

func fetchFiles(c *gin.Context) {
	type FileWithoutContent struct {
		ID         string `json:"id" bson:"_id"`
		CourseName string `json:"courseName" bson:"courseName"`
		Batch      string `json:"batch" bson:"batch"`
		Instructor string `json:"instructor" bson:"instructor"`
		Type       string `json:"type" bson:"type"`
		Remark     string `json:"remark" bson:"remark"`
		Link       string `json:"link" bson:"link"`
	}

	var files []FileWithoutContent

	cursor, err := collection.Find(ctx, bson.M{}, options.Find().SetProjection(bson.M{"fileContent": 0}))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data from database"})
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var file FileWithoutContent
		if err := cursor.Decode(&file); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode data"})
			return
		}

		files = append(files, file)
	}

	if err := cursor.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cursor error"})
		return
	}

	c.JSON(http.StatusOK, files)
}

func downloadFile(c *gin.Context) {
	fileID := c.Param("fileID")

	// Fetch file details from the database using the fileID
	var file File
	err := collection.FindOne(ctx, bson.M{"_id": fileID}).Decode(&file)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Serve the file content
	c.Data(http.StatusOK, "application/octet-stream", file.FileContent)
}

func uploadFaculty(c *gin.Context) {
	var faculty Faculty
	if err := c.ShouldBindJSON(&faculty); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := facultyCollection.InsertOne(ctx, faculty)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store faculty data in database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Faculty data uploaded successfully"})
}

func fetchFaculty(c *gin.Context) {
	var faculties []Faculty

	cursor, err := facultyCollection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch faculty data from database"})
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var faculty Faculty
		if err := cursor.Decode(&faculty); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode faculty data"})
			return
		}
		faculties = append(faculties, faculty)
	}

	if err := cursor.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cursor error"})
		return
	}

	c.JSON(http.StatusOK, faculties)
}

func uploadCourse(c *gin.Context) {
	var course Course // Assuming you have a struct definition for Course similar to Faculty
	if err := c.ShouldBindJSON(&course); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := courseCollection.InsertOne(ctx, course)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store course data in database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Course data uploaded successfully"})
}

func fetchCourses(c *gin.Context) {
	var courses []Course // Assuming you have a struct definition for Course similar to Faculty

	cursor, err := courseCollection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch course data from database"})
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var course Course
		if err := cursor.Decode(&course); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode course data"})
			return
		}
		courses = append(courses, course)
	}

	if err := cursor.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cursor error"})
		return
	}

	c.JSON(http.StatusOK, courses)
}

func register(c *gin.Context) {
	var user UserRegistration
	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if the username already exists in the database
	count, err := registeredUsers.CountDocuments(ctx, bson.M{"username": user.Username})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check username availability"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username already exists"})
		return
	}

	// Generate a random 6-digit OTP
	rand.Seed(time.Now().UnixNano())
	otp := strconv.Itoa(rand.Intn(900000) + 100000) // Generates a random number between 100000 and 999999

	// Hash the password before storing it in the database
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Store user registration data in the database along with the OTP
	_, err = registeredUsers.InsertOne(ctx, bson.M{
		"username":   user.Username,
		"password":   string(hashedPassword),
		"isVerified": false,
		"otp":        otp,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	// Send OTP to the user's email address
	if err := sendVerificationOTP(user.Username, otp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send verification OTP"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully. Please verify your email to activate your account"})
}

func verifyOTP(c *gin.Context) {
	type OTPRequest struct {
		Username string `json:"username" binding:"required"`
		OTP      string `json:"otp" binding:"required"`
	}

	var req OTPRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Query the database to find the user by username
	var dbUser UserRegistration
	err := registeredUsers.FindOne(ctx, bson.M{"username": req.Username}).Decode(&dbUser)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Check if the provided OTP matches the stored OTP
	if req.OTP != dbUser.OTP {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid OTP"})
		return
	}

	// Update the user's verification status to true
	_, err = registeredUsers.UpdateOne(ctx, bson.M{"username": req.Username}, bson.M{"$set": bson.M{"isVerified": true}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify OTP"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP verified successfully"})
}

func sendVerificationOTP(email, otp string) error {
	// SMTP configuration
	smtpHost := "mmtp.iitk.ac.in"
	smtpPort := 25
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	// Sender and recipient email addresses
	from := "EduWise@iitk.ac.in"
	to := email

	// Email content
	subject := "Account Verification OTP"
	body := fmt.Sprintf("Dear User your verification OTP is: %s", otp)

	// Constructing email headers
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/plain; charset=\"utf-8\""
	headers["Content-Transfer-Encoding"] = "base64"

	var msg bytes.Buffer
	for key, value := range headers {
		msg.WriteString(key + ": " + value + "\r\n")
	}
	msg.WriteString("\r\n" + base64.StdEncoding.EncodeToString([]byte(body)))

	// SMTP authentication
	auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost)

	// Sending email using SMTP
	err := smtp.SendMail(fmt.Sprintf("%s:%d", smtpHost, smtpPort), auth, from, []string{to}, msg.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func login(c *gin.Context) {
	var user struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Query the database to find the user by username
	var dbUser UserRegistration
	err := registeredUsers.FindOne(ctx, bson.M{"username": user.Username}).Decode(&dbUser)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Check if the user is verified
	if !dbUser.IsVerified {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account not verified. Please check your email for verification instructions."})
		return
	}

	// Compare the provided password with the hashed password from the database
	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(user.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Generate JWT token
	tokenString, err := generateJWT(user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate JWT token"})
		return
	}

	// Return JWT token to the client
	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func isLoggedIn(c *gin.Context) {
	// Retrieve the JWT token from the Authorization header
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization token"})
		c.Abort()
		return
	}

	// Parse the JWT token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization token"})
		c.Abort()
		return
	}

	// Token is valid, user is logged in
	c.JSON(http.StatusOK, gin.H{"isLoggedIn": true})
}

func isAuthenticated(c *gin.Context) {
	// Retrieve the JWT token from the Authorization header
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization token"})
		c.Abort()
		return
	}

	// Parse the JWT token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization token"})
		c.Abort()
		return
	}

	// Proceed with the request
	c.Next()
}

// Function to generate JWT token
func generateJWT(username string) (string, error) {
	expirationTime := time.Now().Add(7 * 24 * time.Hour) // Token valid for 7 days

	claims := &Claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Modify your main page handler function according to your needs
func mainPageHandler(c *gin.Context) {
	// Retrieve user information from the context
	username, _ := c.Get("username")

	// Your main page logic goes here
	c.JSON(http.StatusOK, gin.H{"message": "Welcome to the main page", "username": username})
}
