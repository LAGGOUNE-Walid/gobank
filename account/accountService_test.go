package account

import "testing"

func TestHashPassword(t *testing.T) {
	hashed, err := HashPassword("secret123")
	if len(hashed) == 0 || err != nil {
		t.Fatal("hashed password is empty")
	}
	if len(hashed) < 60 || hashed[:4] != "$2a$" {
		t.Fatal("hashed password is not a valid bcrypt hash")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	hashed, err := HashPassword("secret123")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if !CheckPasswordHash("secret123", hashed) {
		t.Errorf("check plain password and hash failed")
	}

	if CheckPasswordHash("secret1234", hashed) {
		t.Errorf("check wrong plain password and hash failed")
	}
}

func TestNew(t *testing.T) {
	_, err := New("", "", "", "")
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	acc, err := New("username", "password123", "John", "Doe")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if acc.Username != "username" || acc.Firstname != "John" || acc.Lastname != "Doe" {
		t.Errorf("expected created account with correct fields, got %+v", acc)
	}

	if acc.Password == "password123" {
		t.Fatal("password should be hashed")
	}
}
