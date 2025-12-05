package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"website/config"
	"website/database"
	"website/webserver"

	"github.com/google/uuid"
)

var (
	debugLogging *bool
	mockDB       *bool
)

func init() {
	debugLogging = flag.Bool("debug", false, "Whether to enable debug logging")
	mockDB = flag.Bool("mockdb", false, "Whether to mock a database instead of connecting to the real one")
	flag.Parse()
	config.Load("config.yaml")
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	var mock func()
	if *mockDB {
		mock = loadMockData
	}
	database.Connect(*debugLogging, mock)
	defer database.Close()

	webserver.Start("../webserver")

	fmt.Println("Press Ctrl+C to stop")
	<-ctx.Done()
	fmt.Println("\nShutting down!")
	webserver.WriteCache()
}

func loadMockData() {
	mockUsers := map[string]struct {
		uuid  string
		flags int
	}{
		"Kesu": {
			uuid:  "042f47d5-9a3c-4c98-9a50-c71f38f38ed0",
			flags: database.FlagAll,
		},
		"test123": {
			uuid:  "853c80ef-3c37-49fd-aa49-938b674adae6", // Jeb_
			flags: database.FlagActive,
		},
	}

	for username, mockUserData := range mockUsers {
		pwhash := sha256.Sum256([]byte(username))
		mockUser := database.User{
			Email:    fmt.Sprintf("%s@local", username),
			Username: username,
			Password: hex.EncodeToString(pwhash[:]),
		}
		if err := database.DB.Save(&mockUser).Error; err != nil {
			log.Fatalf("Failed to create mock user: %v", err)
		}
		log.Printf("User ID is %d (%s)", mockUser.ID, mockUser.CreatedAt)
		mockUser.WhitelistEntry = &database.WhitelistEntry{
			UserID: mockUser.ID,
			UUID:   uuid.MustParse(mockUserData.uuid),
			Name:   username,
			Flags:  mockUserData.flags,
		}
		if err := database.DB.Save(&mockUser).Error; err != nil {
			log.Fatalf("Failed to create mock user: %v", err)
		}

		log.Printf("Created Mock User: '%s'", username)
	}
}
