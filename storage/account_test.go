package storage

import (
	"testing"

	"github.com/LAGGOUNE-Walid/gobank/account"
)

func TestCreateAndFetchAccount(t *testing.T) {
	store := SetupTestDB(t)

	acc, err := account.New("test", "test", "test", "test")
	if err != nil {
		t.Fatal("failed to create new account:", err)
	}

	id, err := store.Create(acc)
	if err != nil {
		t.Fatal("failed to create account in DB:", err)
	}

	got, err := store.Find(id)
	if err != nil {
		t.Fatal("failed to find account by ID:", err)
	}

	if got.Username != acc.Username {
		t.Errorf("expected username %s, got %s", acc.Username, got.Username)
	}

	got, err = store.FindByNumber(acc.Number)
	if err != nil {
		t.Fatal("failed to find account by number:", err)
	}

	if got.Number != acc.Number {
		t.Errorf("expected Number %d, got %d", acc.Number, got.Number)
	}

	_, err = store.Find(-1)
	if err == nil {
		t.Fatal("expected error for non-existing ID, got nil")
	}

	_, err = store.FindByNumber(-1)
	if err == nil {
		t.Fatal("expected error for non-existing account number, got nil")
	}
}

func TestAll(t *testing.T) {
	store := SetupTestDB(t)

	accounts, err := store.All(10, 1)
	if err != nil {
		t.Fatal("failed to retrieve accounts:", err)
	}

	if len(accounts.Data) > 0 {
		t.Fatal("expected no accounts, but found some")
	}

	acc1, _ := account.New("test", "test", "test", "test")
	acc2, _ := account.New("test2", "test2", "test2", "test2")
	store.Create(acc1)
	store.Create(acc2)

	accounts, err = store.All(10, 1)
	if err != nil {
		t.Fatal("failed to retrieve accounts:", err)
	}
	if len(accounts.Data) != 2 || accounts.Total != 2 {
		t.Fatalf("expected 2 accounts, got %d", len(accounts.Data))
	}

	accounts, err = store.All(0, 0)
	if err != nil {
		t.Fatal("failed to retrieve accounts with limit 0 and page 0:", err)
	}
	if len(accounts.Data) == 0 {
		t.Fatal("expected accounts, but got none")
	}
}

func TestDelete(t *testing.T) {
	store := SetupTestDB(t)
	acc1Entity, _ := account.New("test1", "test1", "test1", "test1")
	acc2Entity, _ := account.New("test2", "test2", "test2", "test2")
	acc1Id, _ := store.Create(acc1Entity)
	acc2Id, _ := store.Create(acc2Entity)

	store.Delete(acc1Id)

	_, err := store.Find(acc1Id)
	if err == nil {
		t.Errorf("user id %d must not exists in db", acc1Id)
	}

	acc2, _ := store.Find(acc2Id)
	if acc2 == nil {
		t.Errorf("user id %d must exists in db", acc2Id)
	}
}

func TestTransfer(t *testing.T) {
	store := SetupTestDB(t)

	fromAcc, _ := account.New("sender", "sender", "sender", "sender")
	toAcc, _ := account.New("receiver", "receiver", "receiver", "receiver")

	fromID, err := store.Create(fromAcc)
	if err != nil {
		t.Fatal("failed to create from account:", err)
	}
	_, err = store.Create(toAcc)
	if err != nil {
		t.Fatal("failed to create to account:", err)
	}

	_, err = store.db.Exec("UPDATE accounts SET balance = ? WHERE id = ?", 1000, fromID)
	if err != nil {
		t.Fatal("failed to update from account balance:", err)
	}

	err = store.Transfer(toAcc.Number, fromAcc.ID, 300)
	if err != nil {
		t.Fatal("failed to transfer:", err)
	}

	fromEntity, _ := store.Find(fromID)
	toEntity, _ := store.FindByNumber(toAcc.Number)

	if fromEntity.Balance != 700 {
		t.Errorf("expected sender balance 700, got %d", fromEntity.Balance)
	}
	if toEntity.Balance != 300 {
		t.Errorf("expected receiver balance 300, got %d", toEntity.Balance)
	}
}

func TestTransferNotEnoughBalance(t *testing.T) {
	store := SetupTestDB(t)

	fromAcc, _ := account.New("low", "low", "low", "low")
	toAcc, _ := account.New("rich", "rich", "rich", "rich")

	fromID, _ := store.Create(fromAcc)
	_, _ = store.Create(toAcc)

	_, err := store.db.Exec("UPDATE accounts SET balance = ? WHERE id = ?", 100, fromID)
	if err != nil {
		t.Fatal("failed to update from account balance:", err)
	}

	err = store.Transfer(toAcc.Number, fromAcc.ID, 200)
	if err == nil {
		t.Fatal("expected transfer to fail due to insufficient balance")
	}
}

func TestLogin(t *testing.T) {
	store := SetupTestDB(t)

	acc, _ := account.New("test", "test", "test", "test")
	_, err := store.Create(acc)
	if err != nil {
		t.Fatal("failed to create account:", err)
	}

	accountEntity, err := store.Login("test", "test")
	if err != nil {
		t.Fatal("failed to log in with correct credentials:", err)
	}
	if accountEntity.Username != acc.Username {
		t.Fatalf("expected username %s, got %s", acc.Username, accountEntity.Username)
	}

	_, err = store.Login("test", "wrongpassword")
	if err == nil {
		t.Fatal("expected error for incorrect password, got nil")
	}
}
