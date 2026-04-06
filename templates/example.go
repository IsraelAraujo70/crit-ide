package main

import (
	"fmt"
	"strings"
)

// User represents a registered user in the system.
type User struct {
	Name  string
	Email string
	Age   int
	Admin bool
}

// NewUser creates a new User with validation.
func NewUser(name, email string, age int) (*User, error) {
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}
	if age < 0 || age > 150 {
		return nil, fmt.Errorf("invalid age: %d", age)
	}
	return &User{
		Name:  strings.TrimSpace(name),
		Email: email,
		Age:   age,
		Admin: false,
	}, nil
}

/*
  Greet returns a greeting message.
  This is a block comment spanning
  multiple lines.
*/
func (u *User) Greet() string {
	if u.Admin {
		return fmt.Sprintf("Welcome back, admin %s!", u.Name)
	}
	return fmt.Sprintf("Hello, %s! You are %d years old.", u.Name, u.Age)
}

func main() {
	users := make([]*User, 0, 10)
	names := []string{"Alice", "Bob", "Charlie"}

	for i, name := range names {
		user, err := NewUser(name, name+"@example.com", 25+i*5)
		if err != nil {
			panic(err)
		}
		users = append(users, user)
	}

	users[0].Admin = true

	for _, u := range users {
		fmt.Println(u.Greet())
	}

	count := len(users)
	fmt.Printf("Total users: %d\n", count)
}
