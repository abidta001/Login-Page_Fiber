package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/template/html/v2"
)

var store *session.Store

func main() {
	// Initialize Go html template engine (v2 version)
	engine := html.New("./views", ".html")

	// Create a new session store
	store = session.New()

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	// Middleware
	app.Use(logger.New())

	// Middleware to prevent caching
	app.Use(func(c *fiber.Ctx) error {
		c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Set("Pragma", "no-cache")
		c.Set("Expires", "0")
		return c.Next()
	})

	// Middleware to check login status
	app.Use(func(c *fiber.Ctx) error {
		if c.Path() != "/login" && c.Path() != "/logout" {
			sess, err := store.Get(c)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).SendString("Error retrieving session")
			}

			username := sess.Get("username")
			if username == nil {
				return c.Redirect("/login")
			}
		}
		return c.Next()
	})

	// Routes
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/home")
	})

	app.Get("/login", func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error retrieving session")
		}

		// If the user is already logged in, redirect them to the home page
		if sess.Get("username") != nil {
			return c.Redirect("/home")
		}
		app.Get("/home", func(c *fiber.Ctx) error {
			sess, err := store.Get(c)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).SendString("Error retrieving session")
			}

			// Retrieve the username from the session
			username := sess.Get("username")
			if username == nil {
				return c.Redirect("/login")
			}

			// Pass the username to the template
			return c.Render("home", fiber.Map{
				"Username": username,
			})
		})

		// Otherwise, render the login page
		return c.Render("login", fiber.Map{
			"Error": "",
		})
	})

	app.Post("/login", func(c *fiber.Ctx) error {
		username := c.FormValue("username")
		password := c.FormValue("password")

		if username == "" || password == "" {
			return c.Render("login", fiber.Map{
				"Error": "Username and password are required.",
			})
		}

		// Simple hardcoded authentication check
		if username == "Abid" && password == "0000" {
			sess, err := store.Get(c)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).SendString("Error creating session")
			}

			sess.Set("username", username)
			sess.Save()

			// Redirect to home after successful login
			return c.Redirect("/home")
		}

		// If login fails, return to the login page with an error message
		return c.Render("login", fiber.Map{
			"Error": "Invalid username or password.",
		})
	})

	app.Get("/home", func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error retrieving session")
		}

		username := sess.Get("username")
		if username == nil {
			return c.Redirect("/login")
		}

		return c.Render("home", fiber.Map{
			"Username": username,
		})
	})

	app.Post("/logout", func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error retrieving session")
		}

		sess.Destroy()

		// Redirect to login page after logout
		return c.Redirect("/login")
	})

	// Start server
	log.Fatal(app.Listen(":8000"))
}
