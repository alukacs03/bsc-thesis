package controllers

import (
	"fmt"
	"gluon-api/database"
	"gluon-api/models"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var secretKey = func() string {
	s := os.Getenv("SECRET_KEY")
	if s == "" {
		// default for development only; set SECRET_KEY in environment in production
		s = "default_secret"
	}
	return s
}()

func AddDemoUser() {
	var existingUser models.User
	if err := database.DB.Where("email = ?", "admin@example.com").First(&existingUser).Error; err == nil {
		fmt.Println("Demo user already exists:", existingUser)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("Failed to hash password:", err)
		return
	}

	demoUser := models.User{
		Name:      "Admin User",
		Email:     "admin@example.com",
		Password:  hashedPassword,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := database.DB.Create(&demoUser).Error; err != nil {
		fmt.Println("Failed to create demo user:", err)
		return
	}

	fmt.Println("Demo user created successfully:", demoUser)
}

func Hello(c *fiber.Ctx) error {
	return c.SendString("Hello, World!")
}

func Register(c *fiber.Ctx) error {
	fmt.Println("Received a registration request")

	var data map[string]string
	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	var existingUser models.User
	if err := database.DB.Where("email = ?", data["email"]).First(&existingUser).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "User with this email already exists",
		})
	}

	var existingRequest models.UserRegistrationRequest
	if err := database.DB.Where("email = ?", data["email"]).First(&existingRequest).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "A registration request with this email already exists. Wait for approval.",
		})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(data["password"]), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to hash password",
		})
	}

	registerRequest := models.UserRegistrationRequest{
		Email:    data["email"],
		FullName: data["name"],
		Password: hashedPassword,
		Status:   "pending",
	}

	if err := database.DB.Create(&registerRequest).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create registration request",
		})
	}

	/*
		user := models.User{
			Name:     data["name"],
			Email:    data["email"],
			Password: hashedPassword,
		}

		if err := database.DB.Create(&user).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create user",
			})
		}*/
	return c.JSON(fiber.Map{
		"message": "User Registration Request Created Successfully",
	})
}

func ModifyUserRegistration(c *fiber.Ctx) error {
	fmt.Println("Modifying user registration request")

	var data map[string]string
	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	token := c.Locals("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	sub := claims["sub"].(string)
	userID, _ := strconv.Atoi(sub)

	fmt.Println("Admin user:", userID)

	if data["status"] != "approved" && data["status"] != "rejected" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid status value",
		})
	}

	var registrationRequest models.UserRegistrationRequest
	if err := database.DB.Where("id = ?", data["request_id"]).First(&registrationRequest).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Registration request not found",
		})
	}

	uid := uint(userID)

	if data["status"] == "approved" {
		registrationRequest.Status = "approved"
		now := time.Now()
		registrationRequest.ApprovedAt = &now
		registrationRequest.ApprovedByID = &uid
		registrationRequest.ApprovedBy = &models.User{ID: uid}

		user := models.User{
			Name:      registrationRequest.FullName,
			Email:     registrationRequest.Email,
			Password:  registrationRequest.Password,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := database.DB.Create(&user).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create user. Error: " + err.Error(),
			})
		}

		registrationRequest.ConvertedUserID = &user.ID

	} else if data["status"] == "rejected" {
		registrationRequest.Status = "rejected"
		now := time.Now()
		registrationRequest.RejectedAt = &now
		registrationRequest.RejectedByID = &uid
		registrationRequest.RejectedBy = &models.User{ID: uid}
		registrationRequest.RejectionReason = data["rejection_reason"]
	}

	if err := database.DB.Save(&registrationRequest).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update registration request",
		})
	}

	return c.JSON(fiber.Map{
		"message": "User Registration Request Modified Successfully",
	})
}

func ListUserRegRequests(c *fiber.Ctx) error {
	fmt.Println("Listing user registration requests")

	var requests []models.UserRegistrationRequest
	if err := database.DB.Preload("ApprovedBy").Preload("RejectedBy").Find(&requests).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch registration requests",
		})
	}

	return c.JSON(requests)
}

func DeleteUser(c *fiber.Ctx) error {
	fmt.Println("Deleting user")

	var data map[string]string
	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	if data["user_id"] == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user_id is required",
		})
	}

	token := c.Locals("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	sub := claims["sub"].(string)

	if sub == data["user_id"] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Users cannot delete themselves",
		})
	}

	var user models.User
	if err := database.DB.Where("id = ?", data["user_id"]).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	if err := database.DB.Delete(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete user",
		})
	}

	return c.JSON(fiber.Map{
		"message": fmt.Sprintf("User %s deleted successfully", user.Email),
	})
}

func Login(c *fiber.Ctx) error {
	fmt.Println("Received a login request")

	var data map[string]string
	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	var user models.User
	database.DB.Where("email = ?", data["email"]).First(&user)

	if user.ID == 0 {
		fmt.Println("user not found")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid email or password",
		})
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(data["password"]))
	if err != nil {
		fmt.Println("Invalid password: ", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid email or password",
		})
	}

	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": strconv.Itoa(int(user.ID)),
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	token, err := claims.SignedString([]byte(secretKey))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    token,
		Expires:  time.Now().Add(time.Hour * 24),
		HTTPOnly: true,
		Secure:   false,
	}

	c.Cookie(&cookie)

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"message": "Login successful",
	})

}

func User(c *fiber.Ctx) error {
	fmt.Println("Fetching user info")

	cookie := c.Cookies("jwt")

	token, err := jwt.ParseWithClaims(cookie, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil || token == nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	claims, ok := token.Claims.(*jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to parse claims",
		})
	}

	id, _ := strconv.Atoi((*claims)["sub"].(string))

	user := models.User{ID: uint(id)}
	database.DB.Where("id = ?", id).First(&user)

	return c.JSON(user)
}

func Logout(c *fiber.Ctx) error {
	fmt.Println("Logging out user")

	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   false,
	}
	c.Cookie(&cookie)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Logout successful",
	})
}
