package database

import (
	"github.com/supabase-community/supabase-go"
	"os"
	"sync"
)

var (
	client *supabase.Client
	once   sync.Once
)

type Database struct {
	Client *supabase.Client
}

func InitDB() (*Database, error) {
	var err error
	once.Do(func() {
		supabaseUrl := os.Getenv("SUPABASE_URL")
		supabaseKey := os.Getenv("SUPABASE_SERVICE_ROLE")

		client, err = supabase.NewClient(supabaseUrl, supabaseKey, nil)
	})

	if err != nil {
		return nil, err
	}

	return &Database{Client: client}, nil
}

func GetDB() *Database {
	db, err := InitDB()
	if err != nil {
		panic(err)
	}
	return db
}
