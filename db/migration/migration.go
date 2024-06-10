package migration

import (
	"errors"
	"github.com/Montheankul-K/jod-jod/db"
	"github.com/Montheankul-K/jod-jod/domains/transaction"
	"github.com/Montheankul-K/jod-jod/domains/user"
)

func Migrate(db db.DB) error {
	err := db.Connect().AutoMigrate(&user.Users{}, &transaction.Transaction{})
	if err != nil {
		return errors.New("cannot migrate database")
	}
	return nil
}
