package main

import (
	"backend/app/auth"
	"backend/app/models"
	"backend/app/utils"
	"flag"
	"log"
)

func main() {
	email := flag.String("u", "", "user's email")
	pass := flag.String("p", "", "user's password")
	isAdmin := flag.Bool("admin", false, "user should be an admin")
	flag.Parse()

	if *email == "" {
		log.Fatal("Please provide user's email")
	}

	if *pass == "" {
		log.Fatal("Please provide user's password")
	}

	db, err := utils.Init_db()
	if err != nil {
		log.Fatal(err)
	}

	if sqlDB, sqlErr := db.DB(); sqlErr == nil {
		defer sqlDB.Close()
	}

	userAuth := auth.NewUserAuth(db)

	role := models.AnnotatorRole
	if *isAdmin {
		role = models.AdminRole
	}

	if _, userErr := userAuth.CreateUser(*email, *pass, role); err != nil {
		log.Fatal(userErr)
	}

	log.Println("User created!")
}
