package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LAGGOUNE-Walid/gobank/account"
	"github.com/LAGGOUNE-Walid/gobank/storage"
)

func TestCreateAndGettingAccount(t *testing.T) {
	store := storage.SetupTestDB(t)
	server := NewServer(":8080", store)
	handler := server.Routes()

	t.Run("Empty database returns empty list", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/account?page=1", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var accounts account.Response
		if err := json.NewDecoder(w.Body).Decode(&accounts); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(accounts.Data) != 0 {
			t.Errorf("expected 0 accounts, got %d", len(accounts.Data))
		}
	})

	t.Run("Create account and retrieve it", func(t *testing.T) {
		body := account.CreateRequest{
			Username:  "foo",
			Password:  "bar",
			Firstname: "foo",
			Lastname:  "bar",
		}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/account", bytes.NewReader(jsonBody))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d", w.Code)
		}

		var resp account.CreateResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.ID == 0 {
			t.Errorf("expected non-zero ID")
		}

		getReq := httptest.NewRequest("GET", "/account?page=1", nil)
		getW := httptest.NewRecorder()

		handler.ServeHTTP(getW, getReq)

		if getW.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", getW.Code)
		}

		var accounts account.Response
		if err := json.NewDecoder(getW.Body).Decode(&accounts); err != nil {
			t.Fatalf("failed to decode accounts list: %v", err)
		}

		if len(accounts.Data) == 0 {
			t.Fatalf("expected at least 1 account, got 0")
		}

		found := false
		for _, acc := range accounts.Data {
			if acc.ID == resp.ID {
				found = true
			}
		}
		if !found {
			t.Errorf("created account with ID %d not found in account list", resp.ID)
		}
	})

	t.Run("Getting pagination after the last page will return empty accounts list", func(t *testing.T) {
		body := account.CreateRequest{
			Username:  "foo",
			Password:  "bar",
			Firstname: "foo",
			Lastname:  "bar",
		}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/account", bytes.NewReader(jsonBody))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d", w.Code)
		}

		var resp account.CreateResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.ID == 0 {
			t.Errorf("expected non-zero ID")
		}

		getReq := httptest.NewRequest("GET", "/account?page=100", nil)
		getW := httptest.NewRecorder()

		handler.ServeHTTP(getW, getReq)

		if getW.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", getW.Code)
		}

		var accounts account.Response
		if err := json.NewDecoder(getW.Body).Decode(&accounts); err != nil {
			t.Fatalf("failed to decode accounts list: %v", err)
		}

		if len(accounts.Data) > 0 {
			t.Fatalf("expected 0 account, got %d", len(accounts.Data))
		}

	})

	t.Run("Passing invalid limits will prevent from crushing", func(t *testing.T) {
		body := account.CreateRequest{
			Username:  "foo",
			Password:  "bar",
			Firstname: "foo",
			Lastname:  "bar",
		}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/account", bytes.NewReader(jsonBody))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d", w.Code)
		}

		var resp account.CreateResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.ID == 0 {
			t.Errorf("expected non-zero ID")
		}

		getReq := httptest.NewRequest("GET", "/account?page=1&limit=-2381903812312", nil)
		getW := httptest.NewRecorder()

		handler.ServeHTTP(getW, getReq)

		if getW.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", getW.Code)
		}

		var accounts account.Response
		if err := json.NewDecoder(getW.Body).Decode(&accounts); err != nil {
			t.Fatalf("failed to decode accounts list: %v", err)
		}

		if len(accounts.Data) == 0 {
			t.Fatal("expected at least one account got 0")
		}

	})
}

func TestLogin(t *testing.T) {
	store := storage.SetupTestDB(t)
	server := NewServer(":8080", store)

	for i := 0; i < 3; i++ {
		acc, _ := account.New(
			fmt.Sprintf("user%d", i),
			"password",
			"First",
			"Last",
		)
		_, err := store.Create(acc)
		if err != nil {
			t.Fatalf("failed to create test account: %v", err)
		}
	}

	t.Run("Loggin with correct data", func(t *testing.T) {
		loginRequest := account.LoginRequest{Username: "user1", Password: "password"}
		jsonBody, _ := json.Marshal(loginRequest)
		req := httptest.NewRequest("POST", "/login", bytes.NewReader(jsonBody))
		w := httptest.NewRecorder()

		handler := server.Routes()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var loginResponse account.LoginResponse
		if err := json.NewDecoder(w.Body).Decode(&loginResponse); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(loginResponse.Token) == 0 {
			t.Fatalf("failed to generate token")
		}
	})

	t.Run("Loggin with false data", func(t *testing.T) {
		loginRequest := account.LoginRequest{Username: "user1", Password: "falsepassword"}
		jsonBody, _ := json.Marshal(loginRequest)
		req := httptest.NewRequest("POST", "/login", bytes.NewReader(jsonBody))
		w := httptest.NewRecorder()

		handler := server.Routes()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			fmt.Println(w.Body)
			t.Fatalf("expected status 400, got %d", w.Code)
		}

		loginRequest = account.LoginRequest{Username: "falseusername", Password: "password"}
		jsonBody, _ = json.Marshal(loginRequest)
		req = httptest.NewRequest("POST", "/login", bytes.NewReader(jsonBody))
		w = httptest.NewRecorder()

		handler = server.Routes()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			fmt.Println(w.Body)
			t.Fatalf("expected status 400, got %d", w.Code)
		}
	})

}

func TestGettingAccount(t *testing.T) {
	store := storage.SetupTestDB(t)
	server := NewServer(":8080", store)

	for i := 0; i < 3; i++ {
		acc, _ := account.New(
			fmt.Sprintf("user%d", i),
			"password",
			"First",
			"Last",
		)
		_, err := store.Create(acc)
		if err != nil {
			t.Fatalf("failed to create test account: %v", err)
		}
	}

	req := httptest.NewRequest("GET", fmt.Sprintf("/account/%d", 1), nil)
	req.SetPathValue("id", "1")

	tokenString, _ := createToken(1, "foo")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenString))

	w := httptest.NewRecorder()

	handler := server.Routes()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var accountResponse account.Entity
	if err := json.NewDecoder(w.Body).Decode(&accountResponse); err != nil {
		t.Fatalf("failed to decode response: %v+ with error of %v", w.Body, err)
	}

	if accountResponse.ID != 1 {
		t.Errorf("expected account id  got %d", accountResponse.ID)
	}
}

func TestDeletingAccount(t *testing.T) {
	store := storage.SetupTestDB(t)
	server := NewServer(":8080", store)
	for i := 0; i < 3; i++ {
		acc, _ := account.New(
			fmt.Sprintf("user%d", i),
			"password",
			"First",
			"Last",
		)
		_, err := store.Create(acc)
		if err != nil {
			t.Fatalf("failed to create test account: %v", err)
		}
	}
	tokenString, _ := createToken(1, "user1")

	t.Run("Authenticated user can't delete others account", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/account/2", nil)
		req.SetPathValue("id", "2")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenString))

		w := httptest.NewRecorder()

		handler := server.Routes()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", w.Code)
		}

		_, err := store.Find(2)
		if err != nil {
			t.Fatal("user deleted from database when it shouldn't ")
		}
	})

	t.Run("Authenticated user can delete his account", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/account/1", nil)
		req.SetPathValue("id", "1")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenString))

		w := httptest.NewRecorder()

		handler := server.Routes()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d", w.Code)
		}

		_, err := store.Find(1)
		if err == nil {
			t.Fatal("user not deleted from database")
		}
	})

}

func TestTransfer(t *testing.T) {
	store := storage.SetupTestDB(t)
	server := NewServer(":8080", store)
	for i := 0; i < 3; i++ {
		acc, _ := account.New(
			fmt.Sprintf("user%d", i),
			"password",
			"First",
			"Last",
		)
		_, err := store.Create(acc)
		if err != nil {
			t.Fatalf("failed to create test account: %v", err)
		}
	}
	targetAccount, _ := store.Find(2)
	tokenString, _ := createToken(1, "user1")

	t.Run("it blocks when the ammount is 0", func(t *testing.T) {
		body := account.TransferRequest{To: targetAccount.Number, Ammout: 0}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/transfer", bytes.NewReader(jsonBody))
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenString))
		w := httptest.NewRecorder()

		handler := server.Routes()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("it blocks when the target number is invalid", func(t *testing.T) {
		body := account.TransferRequest{To: -1, Ammout: 1000}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/transfer", bytes.NewReader(jsonBody))
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenString))
		w := httptest.NewRecorder()

		handler := server.Routes()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", w.Code)
		}
	})
	t.Run("it blocks the sender have no credit", func(t *testing.T) {
		body := account.TransferRequest{To: targetAccount.Number, Ammout: 1000}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/transfer", bytes.NewReader(jsonBody))
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenString))
		w := httptest.NewRecorder()

		handler := server.Routes()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("it blocks the target do not exists", func(t *testing.T) {
		acc, _ := account.New(
			"rich user",
			"password",
			"First",
			"Last",
		)
		acc.Balance = 2000
		richAccountId, _ := store.Create(acc)
		tokenString, _ := createToken(richAccountId, acc.Username)

		body := account.TransferRequest{To: 2313122321312, Ammout: 1000}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/transfer", bytes.NewReader(jsonBody))
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenString))
		w := httptest.NewRecorder()

		handler := server.Routes()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("transfer the money", func(t *testing.T) {
		acc, _ := account.New(
			"rich user",
			"password",
			"First",
			"Last",
		)
		acc.Balance = 2000
		richAccountId, _ := store.Create(acc)
		tokenString, _ := createToken(richAccountId, acc.Username)

		body := account.TransferRequest{To: targetAccount.Number, Ammout: 1000}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/transfer", bytes.NewReader(jsonBody))
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenString))
		w := httptest.NewRecorder()

		handler := server.Routes()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d", w.Code)
		}

		targertAccount, _ := store.Find(2)
		fromAccount, _ := store.Find(richAccountId)

		if targertAccount.Balance != 1000 {
			t.Fatalf("target account id %d balance is not what the sender sent", targertAccount.ID)
		}

		if fromAccount.Balance != 1000 {
			t.Fatalf("sender account id %d balance is not what the sender sent", fromAccount.ID)
		}

	})
}

func BenchmarkCreateAccount(b *testing.B) {
	store := storage.SetupTestDB(nil)
	server := NewServer(":8080", store)
	handler := server.Routes()

	for i := 0; i < b.N; i++ {
		body := account.CreateRequest{
			Username:  fmt.Sprintf("user%d", i),
			Password:  "password",
			Firstname: "First",
			Lastname:  "Last",
		}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/account", bytes.NewReader(jsonBody))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)
	}
}

func BenchmarkListAccounts(b *testing.B) {
	store := storage.SetupTestDB(nil)
	server := NewServer(":8080", store)
	handler := server.Routes()

	for i := 0; i < 100; i++ {
		acc, _ := account.New(
			fmt.Sprintf("user%d", i),
			"password",
			"First",
			"Last",
		)
		store.Create(acc)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/account?page=1", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

func BenchmarkLogin(b *testing.B) {
	store := storage.SetupTestDB(nil)
	server := NewServer(":8080", store)
	handler := server.Routes()

	acc, _ := account.New("benchmarkuser", "password", "First", "Last")
	store.Create(acc)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		loginRequest := account.LoginRequest{Username: "benchmarkuser", Password: "password"}
		jsonBody, _ := json.Marshal(loginRequest)
		req := httptest.NewRequest("POST", "/login", bytes.NewReader(jsonBody))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

func BenchmarkTransfer(b *testing.B) {
	store := storage.SetupTestDB(nil)
	server := NewServer(":8080", store)
	handler := server.Routes()

	richAcc, _ := account.New("rich", "password", "First", "Last")
	richAcc.Balance = 1_000_000
	richID, _ := store.Create(richAcc)

	recipient, _ := account.New("receiver", "password", "First", "Last")
	store.Create(recipient)

	tokenString, _ := createToken(richID, richAcc.Username)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		body := account.TransferRequest{To: recipient.Number, Ammout: 10}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/transfer", bytes.NewReader(jsonBody))
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenString))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

func BenchmarkParallelTransfer(b *testing.B) {
	store := storage.SetupTestDB(nil)
	server := NewServer(":8080", store)
	handler := server.Routes()

	richAcc, _ := account.New("rich", "password", "First", "Last")
	richAcc.Balance = 1_000_000
	richID, _ := store.Create(richAcc)

	recipient, _ := account.New("receiver", "password", "First", "Last")
	store.Create(recipient)

	tokenString, _ := createToken(richID, richAcc.Username)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				body := account.TransferRequest{To: recipient.Number, Ammout: 10}
				jsonBody, _ := json.Marshal(body)
				req := httptest.NewRequest("POST", "/transfer", bytes.NewReader(jsonBody))
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenString))
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, req)
			}
		})
	}
}
