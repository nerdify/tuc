package tuc

// Card is an individual's card for an user.
type Card struct {
	Balance float64 `json:"balance"`
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Number  string  `json:"number"`
	UserID  string  `json:"-" dynamodbav:"u_id"`
}

// CardService represents a service for managing cards.
type CardService interface {
	List(userID string) ([]Card, error)
	Get(userID, cardID string) (*Card, error)
	Create(card *Card) error
	Update(userID, cardID string, balance float64) (*Card, error)
	Delete(userID, cardID string) error
}

// LoginRequest is a login request for a user.
type LoginRequest struct {
	RequestToken      string `json:"request_token"`
	UserID            string `json:"-" dynamodbav:"u_id"`
	VerificationToken string `json:"verification_token"`
	Verified          bool   `json:"verified"`
}

// LoginRequestService represents a service for managing login requests.
type LoginRequestService interface {
	Create(request *LoginRequest) error
	Delete(email, code string) error

	Verify(email, token string) error
}

// User is an individual's account on Saldo TUC.
type User struct {
	ID string `json:"id"`
}

// UserService represents a service for managing users.
type UserService interface {
	Find(email string) (*User, error)
	Create(user *User) error
}
