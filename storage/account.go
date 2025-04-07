package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/LAGGOUNE-Walid/gobank/account"
)

type AccountStore interface {
	All(limit int, page int) (account.Response, error)
	Find(id int) (account *account.Entity, err error)
	Create(*account.Entity) (int, error)
	Delete(id int) error
	Transfer(to int, from int, ammout int) error
	Login(username string, password string) (*account.Entity, error)
}

func (store *SqliteStore) All(limit int, page int) (response account.Response, err error) {
	if limit <= 0 || limit >= 50 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * limit

	var total int
	if err := store.db.QueryRow("SELECT COUNT(*) FROM accounts").Scan(&total); err != nil {
		return account.Response{}, err
	}

	rows, err := store.db.Query(`
		SELECT id, username, firstname, lastname, balance, number, created_at
		FROM accounts
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return account.Response{}, err
	}
	defer rows.Close()

	accounts := []account.Entity{}
	for rows.Next() {
		var accountEntity account.Entity
		if err := rows.Scan(
			&accountEntity.ID,
			&accountEntity.Username,
			&accountEntity.Firstname,
			&accountEntity.Lastname,
			&accountEntity.Balance,
			&accountEntity.Number,
			&accountEntity.CreatedAt,
		); err != nil {
			return account.Response{}, err
		}
		accounts = append(accounts, accountEntity)
	}

	return account.Response{
		Data:  accounts,
		Page:  page,
		Limit: limit,
		Total: total,
	}, nil
}

func (store *SqliteStore) Login(username string, password string) (*account.Entity, error) {
	var accountEntity account.Entity
	query := `SELECT id, username, password FROM accounts WHERE username = ? LIMIT 1`

	row := store.db.QueryRow(query, username)
	err := row.Scan(&accountEntity.ID, &accountEntity.Username, &accountEntity.Password)
	if err != nil {
		return nil, err
	}
	if !account.CheckPasswordHash(password, accountEntity.Password) {
		return nil, fmt.Errorf("invalid password")
	}
	return &accountEntity, nil
}

func (store *SqliteStore) Find(id int) (*account.Entity, error) {
	var accountEntity account.Entity
	query := `SELECT id, username, firstname, lastname, number, balance, created_at FROM accounts WHERE id = ? LIMIT 1`
	row := store.db.QueryRow(query, id)
	err := row.Scan(&accountEntity.ID, &accountEntity.Username, &accountEntity.Firstname, &accountEntity.Lastname, &accountEntity.Number, &accountEntity.Balance, &accountEntity.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no record by id %d", id)
		}
		return nil, err
	}

	return &accountEntity, nil
}

func (store *SqliteStore) FindByNumber(number int) (*account.Entity, error) {
	var accountEntity account.Entity
	query := `SELECT id, username, firstname, lastname, number, balance, created_at FROM accounts WHERE number = ? LIMIT 1`
	row := store.db.QueryRow(query, number)
	err := row.Scan(&accountEntity.ID, &accountEntity.Username, &accountEntity.Firstname, &accountEntity.Lastname, &accountEntity.Number, &accountEntity.Balance, &accountEntity.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no record by account number %d", number)
		}
		return nil, err
	}

	return &accountEntity, nil
}

func (store *SqliteStore) Create(acc *account.Entity) (int, error) {
	result, err := store.db.Exec("INSERT INTO accounts(username, firstname, lastname, number, balance, password, created_at) values (?,?,?,?,?,?,?)", acc.Username, acc.Firstname, acc.Lastname, acc.Number, acc.Balance, acc.Password, acc.CreatedAt)
	if err != nil {
		return 0, err
	}
	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	acc.ID = int(lastInsertID)
	return acc.ID, nil
}

func (store *SqliteStore) Delete(id int) error {
	query := `DELETE FROM accounts where id = ?`
	if _, err := store.db.Exec(query, id); err != nil {
		return err
	}
	return nil
}

func (store *SqliteStore) Transfer(to int, from int, ammount int) error {
	tx, err := store.db.BeginTx(context.Background(), nil)

	defer tx.Rollback()

	if err != nil {
		return err
	}
	toAccount, err := store.FindByNumber(to)
	if err != nil {
		return err
	}

	fromAccount, err := store.Find(from)
	if err != nil {
		return err
	}

	if fromAccount.Balance-ammount < 0 {
		return fmt.Errorf("you don't have enough credit to do this transaction")
	}

	_, err = tx.ExecContext(context.Background(), "UPDATE accounts SET balance = balance + ? WHERE number = ?",
		ammount, toAccount.Number)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(context.Background(), "UPDATE accounts SET balance = balance - ? WHERE number = ?",
		ammount, fromAccount.Number)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}
