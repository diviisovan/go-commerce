package user

import "fmt"

// User represents a customer in the eCommerce system
type User struct {
	ID       int
	Name     string
	Email    string
	Address  string
	Phone    string
	IsActive bool
}

// Manager manages users in the system
type Manager struct {
	Users  map[int]*User
	NextID int
}

// NewManager creates a new user manager
func NewManager() *Manager {
	return &Manager{
		Users:  make(map[int]*User),
		NextID: 1,
	}
}

// RegisterUser creates a new user account
func (um *Manager) RegisterUser(name, email, address, phone string) *User {
	user := &User{
		ID:       um.NextID,
		Name:     name,
		Email:    email,
		Address:  address,
		Phone:    phone,
		IsActive: true,
	}
	um.Users[um.NextID] = user
	um.NextID++
	return user
}

// GetUser retrieves a user by ID
func (um *Manager) GetUser(id int) (*User, bool) {
	user, exists := um.Users[id]
	return user, exists
}

// UpdateUser updates user information
func (um *Manager) UpdateUser(id int, name, email, address, phone string) error {
	user, exists := um.Users[id]
	if !exists {
		return fmt.Errorf("user with ID %d not found", id)
	}

	if name != "" {
		user.Name = name
	}
	if email != "" {
		user.Email = email
	}
	if address != "" {
		user.Address = address
	}
	if phone != "" {
		user.Phone = phone
	}

	return nil
}

// DisplayUserInfo displays user information
func (u *User) DisplayUserInfo() {
	fmt.Printf("\n=== User Information ===\n")
	fmt.Printf("ID:      %d\n", u.ID)
	fmt.Printf("Name:    %s\n", u.Name)
	fmt.Printf("Email:   %s\n", u.Email)
	fmt.Printf("Address: %s\n", u.Address)
	fmt.Printf("Phone:   %s\n", u.Phone)
	fmt.Printf("Status:  %s\n", map[bool]string{true: "Active", false: "Inactive"}[u.IsActive])
}
