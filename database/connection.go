package database

import (
	"fmt"
	"log"

	// mysql driver used for database
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

type connectionConfig struct {
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
}

func Connect(debug bool, mock func()) {
	log.Println("Connecting to database...")

	var (
		err       error
		dialector gorm.Dialector
	)
	if mock != nil {
		dialector = connectSQLite("file::memory:?cache=shared")
		defer func() {
			if err = DB.AutoMigrate(User{}, WhitelistEntry{}, Marker{}); err != nil {
				log.Fatalf("Could not migrate database: %v", err)
			}
			mock()
		}()
	} else if viper.IsSet("mysql") {
		dialector = connectMySQL()
	} else if viper.IsSet("sqlite") {
		dialector = connectSQLite()
	} else {
		log.Fatalf("No database connection configuration found")
	}

	DB, err = gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		log.Fatalf("Could not open Gorm database connection: %v", err)
	}
	if debug {
		DB = DB.Debug()
	}

	log.Println("Connected to database")
}

// Close closes the database and prevents new queries from starting.
// Close then waits for all queries that have started processing on the server
// to finish.
//
// It is rare to Close a DB, as the DB handle is meant to be
// long-lived and shared between many goroutines.
func Close() {
	if underlyingDB, err := DB.DB(); err == nil {
		underlyingDB.Close()
		log.Println("Closed connection to database")
	} else {
		log.Printf("Could not close database connection: %v", err)
	}
}

func connectMySQL() gorm.Dialector {
	// setting default values
	config := connectionConfig{
		Host: "localhost",
		Port: 3306,
	}

	err := viper.UnmarshalKey("mysql", &config)
	if err != nil {
		log.Fatalf("Could not read msql connection data from config: %v", err)
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", config.User, config.Password, config.Host, config.Port, config.Database)
	return mysql.Open(dsn)
}

func connectSQLite(name ...string) gorm.Dialector {
	if len(name) == 0 || name[0] == "" {
		name = []string{viper.GetString("sqlite")}
	}
	return sqlite.Open(name[0])
}
