package main

import (
	"bytes"
	"fmt"
	"regexp"

	_ "github.com/lib/pq"
	phonedb "phone-number-normalizer/db"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "postgres_phone"
)

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user= %s password=%s sslmode=disable", host, port, user, password)
	must(phonedb.Restet("postgres", psqlInfo, dbname))

	psqlInfo = fmt.Sprintf("%s dbname=%s", psqlInfo, dbname)
	must(phonedb.Migrate("postgres", psqlInfo))

	db, err := phonedb.Open("postgres", psqlInfo)
	must(err)
	defer db.Close()

	err = db.Seed()
	must(err)

	phoneNumbers, err := db.GetAllPhones()
	must(err)

	procesisngNormalizePhones(db, phoneNumbers)

}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func normalize(phone string) string {
	// Pssible combinations using regex
	// `[\(\) -]`
	// `[^0-9]`
	// "\\d" => one digits
	// "\\D" => if not match all digits
	re := regexp.MustCompile(`[^0-9]`)
	newPhoneNormalize := re.ReplaceAllString(phone, "")
	return newPhoneNormalize
}

func normalizeWithBuffer(phone string) string {
	var buf bytes.Buffer

	for _, char := range phone {
		if char >= '0' && char <= '9' {
			buf.WriteRune(char)
		}
	}
	return buf.String()
}

func procesisngNormalizePhones(db *phonedb.DB, phones []phonedb.Phone) {
	for _, phone := range phones {
		fmt.Printf("Working on... %+v\n", phone)
		number := normalize(phone.Number)
		if number != phone.Number {
			fmt.Println("Updating or removing...", number)
			existing, err := db.FindPhone(number)
			must(err)
			if existing != nil {
				must(db.DeletePhoneNumber(phone.ID))
			} else {
				phone.Number = number
				must(db.UpdatePhoneNumber(phone))
			}
		} else {
			fmt.Println("No changes required.")
		}
	}
}
