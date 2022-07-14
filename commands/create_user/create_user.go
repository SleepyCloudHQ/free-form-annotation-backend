package main

import (
	"backend/app/auth"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
)

func main() {
	db, err := gorm.Open(sqlite.Open("../../test.db"))
	if err != nil {
		log.Fatal(err)
	}

	userAuth := auth.NewUserAuth(db)
	if _, userErr := userAuth.CreateUser("test@test.com", "password"); err != nil {
		log.Fatal(userErr)
	}

	log.Println("User created!")
}
