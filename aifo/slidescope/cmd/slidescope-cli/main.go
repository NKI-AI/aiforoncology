// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"syscall"

	"aifo.dev/aifo/slidescope/internal/datasources/database"
	"aifo.dev/aifo/slidescope/internal/server/auth"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"
)

// Slide represents a slide and its associated masks in the YAML file
type Slide struct {
	ImageName string      `yaml:"image_name"`
	ImagePath string      `yaml:"image_path"`
	MaskPath  interface{} `yaml:"mask_path"` // Can be string or []string
}

// Config represents the YAML configuration file structure
type Config []Slide

// AuthResponse represents the response from the authentication endpoint
type AuthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Error       string `json:"error,omitempty"`
}

// SlideResponse represents the response from creating a slide
type SlideResponse struct {
	SlideID   string `json:"slideId"`
	SlideURI  string `json:"slideUri"`
	SlideName string `json:"slideName"`
}

// MaskResponse represents the response from creating a mask
type MaskResponse struct {
	MaskID  string `json:"maskId"`
	MaskURI string `json:"maskUri"`
	Name    string `json:"name"`
}

// Client represents the HTTP client with authentication state
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new API client
func NewClient(baseURL string) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	client := &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Jar: jar,
		},
	}

	return client, nil
}

// Authenticate logs into the server with the given credentials
func (c *Client) Authenticate(username, password string) error {
	authData := map[string]string{
		"username": username,
		"password": password,
	}

	jsonData, err := json.Marshal(authData)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/v1/auth/login", c.BaseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("authentication failed: %s (status code: %d)", string(body), resp.StatusCode)
	}

	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return err
	}

	if authResp.AccessToken == "" {
		return fmt.Errorf("authentication failed: %s", authResp.Error)
	}

	fmt.Printf("Authenticated successfully\n")
	return nil
}

// CreateSlide creates a new slide
func (c *Client) CreateSlide(name, path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	slideData := map[string]string{
		"slideName": name,
		"slideUri":  absPath,
	}

	jsonData, err := json.Marshal(slideData)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/api/v1/slides", c.BaseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to create slide: %s (status code: %d)", string(body), resp.StatusCode)
	}

	var slideResp SlideResponse
	if err := json.NewDecoder(resp.Body).Decode(&slideResp); err != nil {
		return "", err
	}

	return slideResp.SlideID, nil
}

// CreateMask creates a new mask for a slide
func (c *Client) CreateMask(slideID, name, path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	maskData := map[string]string{
		"maskName": name,
		"maskUri":  absPath,
	}

	jsonData, err := json.Marshal(maskData)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/api/v1/slides/%s/annotations/raster", c.BaseURL, slideID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to create mask: %s (status code: %d)", string(body), resp.StatusCode)
	}

	var maskResp MaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&maskResp); err != nil {
		return "", err
	}

	return maskResp.MaskID, nil
}

// readPassword prompts for a password without echoing it to the terminal
func readPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	password, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // Add a newline after the password input
	if err != nil {
		return "", err
	}
	return string(password), nil
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "slidescope",
		Short: "SlideScope command line interface",
		Long:  `A CLI tool for managing SlideScope slides, masks, and users.`,
	}

	// Import command
	importCmd := &cobra.Command{
		Use:   "import",
		Short: "Import slides and masks into SlideScope",
		Long:  `Import slides and masks into the SlideScope server from a YAML configuration file.`,
		RunE:  runImport,
	}
	importCmd.Flags().StringP("config", "c", "", "Path to the YAML configuration file")
	importCmd.Flags().StringP("server", "s", "http://localhost:3000", "URL of the SlideScope server")
	importCmd.Flags().StringP("username", "u", "", "Username for authentication")
	importCmd.MarkFlagRequired("config")

	// User command group
	userCmd := &cobra.Command{
		Use:   "user",
		Short: "User management commands",
		Long:  `Commands for managing SlideScope users.`,
	}

	// Add user command
	addUserCmd := &cobra.Command{
		Use:   "add [username]",
		Short: "Add a new user",
		Long:  `Add a new user to the SlideScope database.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runAddUser,
	}
	addUserCmd.Flags().StringP("database", "d", "", "Database URL (defaults to DATABASE_URL env var)")

	// Change password command
	passwdCmd := &cobra.Command{
		Use:   "passwd [username]",
		Short: "Change a user's password",
		Long:  `Change the password for an existing SlideScope user.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runPasswd,
	}
	passwdCmd.Flags().StringP("database", "d", "", "Database URL (defaults to DATABASE_URL env var)")

	// Add commands to user command group
	userCmd.AddCommand(addUserCmd, passwdCmd)

	// Add all command groups to root command
	rootCmd.AddCommand(importCmd, userCmd)

	// Execute
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runImport(cmd *cobra.Command, args []string) error {
	configFile, _ := cmd.Flags().GetString("config")
	serverURL, _ := cmd.Flags().GetString("server")
	username, _ := cmd.Flags().GetString("username")

	// Prompt for username if not provided
	if username == "" {
		fmt.Print("Username: ")
		fmt.Scanln(&username)
	}

	// Prompt for password
	password, err := readPassword("Password: ")
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}

	// Read and parse the YAML file
	yamlFile, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read YAML file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		return fmt.Errorf("failed to parse YAML file: %w", err)
	}

	// Create and authenticate the client
	client, err := NewClient(serverURL)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	if err := client.Authenticate(username, password); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Process each slide in the config
	for i, slide := range config {
		fmt.Printf("Processing slide %d/%d: %s\n", i+1, len(config), slide.ImageName)

		// Create the slide
		slideID, err := client.CreateSlide(slide.ImageName, slide.ImagePath)
		if err != nil {
			return fmt.Errorf("failed to create slide '%s': %w", slide.ImageName, err)
		}
		fmt.Printf("Created slide with ID: %s\n", slideID)

		// Process masks
		var maskPaths []string

		switch v := slide.MaskPath.(type) {
		case string:
			if v != "" {
				maskPaths = []string{v}
			}
		case []interface{}:
			for _, maskPath := range v {
				if maskStr, ok := maskPath.(string); ok && maskStr != "" {
					maskPaths = append(maskPaths, maskStr)
				}
			}
		}

		// Create each mask
		for j, maskPath := range maskPaths {
			maskName := fmt.Sprintf("%s_mask_%d", slide.ImageName, j+1)
			maskID, err := client.CreateMask(slideID, maskName, maskPath)
			if err != nil {
				return fmt.Errorf("failed to create mask for slide '%s': %w", slide.ImageName, err)
			}
			fmt.Printf("Created mask with ID: %s\n", maskID)
		}
	}

	fmt.Println("Import completed successfully")
	return nil
}

func runAddUser(cmd *cobra.Command, args []string) error {
	username := args[0]
	dbURL, _ := cmd.Flags().GetString("database")

	// Get database URL from env var if not provided
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
		if dbURL == "" {
			return fmt.Errorf("database URL is required (use --database flag or set DATABASE_URL environment variable)")
		}
	}

	// Prompt for password
	password, err := readPassword("Enter password: ")
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}

	confirmPassword, err := readPassword("Confirm password: ")
	if err != nil {
		return fmt.Errorf("failed to read confirmation password: %w", err)
	}

	if password != confirmPassword {
		return fmt.Errorf("passwords do not match")
	}

	// Create database connection
	ctx := context.Background()
	db, err := database.NewDatabase(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.CloseConnections()

	// Hash the password
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Create the user
	newUser := database.NewUser{
		Username: username,
		Password: hashedPassword,
	}

	err = db.CreateUser(ctx, newUser)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	fmt.Printf("User '%s' created successfully\n", username)
	return nil
}

func runPasswd(cmd *cobra.Command, args []string) error {
	username := args[0]
	dbURL, _ := cmd.Flags().GetString("database")

	// Get database URL from env var if not provided
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
		if dbURL == "" {
			return fmt.Errorf("database URL is required (use --database flag or set DATABASE_URL environment variable)")
		}
	}

	// Prompt for new password
	password, err := readPassword("Enter new password: ")
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}

	if len(password) == 0 {
		return fmt.Errorf("password cannot be empty")
	}

	confirmPassword, err := readPassword("Confirm new password: ")
	if err != nil {
		return fmt.Errorf("failed to read confirmation password: %w", err)
	}

	if password != confirmPassword {
		return fmt.Errorf("passwords do not match")
	}

	// Create database connection
	ctx := context.Background()
	db, err := database.NewDatabase(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.CloseConnections()

	// Hash the password
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update the user's password
	err = db.UpdateUserPassword(ctx, username, hashedPassword)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	fmt.Printf("Password for user '%s' updated successfully\n", username)
	return nil
}
