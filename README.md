# GoBank API

GoBank is a simple banking API written in Go with support for account management, user authentication, and money transfers. It uses SQLite as the database engine and includes database migration support via [golang-migrate](https://github.com/golang-migrate/migrate).

## Features

- User registration and login
- Secure authentication with JWT
- Create, list, and delete bank accounts
- Transfer money between accounts
- Database migrations using `migrate` CLI

## Getting Started

### Prerequisites

- Docker installed
- Git (for cloning the repo)

### Build and Run with Docker

```bash
# Clone the repository
git clone https://github.com/LAGGOUNE-Walid/gobank.git
cd gobank

# Build the Docker image
docker build -t gobank .

# Run the container
docker run -p 8081:8080 -v $(pwd)/db:/app/db gobank
```

The API will be available at `http://localhost:8081`.

## API Endpoints

### Auth

#### Create account
`POST /account`

**Request JSON**:
```json
{
  "username": "johndoe",
  "password": "yourpassword"
}
```

#### Login
`POST /login`

**Request JSON**:
```json
{
  "username": "johndoe",
  "password": "yourpassword"
}
```

**Response**:
```json
{
    "id": 1,
    "token": "TOKEN"
}
```

---

#### List Accounts 
`GET /accounts?page=1`


#### Get single Account  (Auth required)
`DELETE /accounts/{id}`

#### Delete Account  (Auth required)
`DELETE /accounts/{id}`

---

### Transfers (Protected)

#### Create Transfer  
`POST /transfer`

**Request JSON**:
```json
{
    "to" : ACCOUNT_NUMBER,
    "ammount" : 600
}
```