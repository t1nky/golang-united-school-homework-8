package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

type Arguments map[string]string

type User struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func (user User) ToString() (string, error) {
	if user.Id == "" {
		return "", nil
	}
	result, err := json.Marshal(user)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

type Users map[string]User

func (users *Users) Add(user User) error {
	if _, exists := (*users)[user.Id]; exists {
		return fmt.Errorf("Item with id %s already exists", user.Id)
	}
	(*users)[user.Id] = user
	return nil
}
func (users *Users) Remove(id string) error {
	if _, ok := (*users)[id]; !ok {
		return fmt.Errorf("Item with id %s not found", id)
	}
	delete((*users), id)
	return nil
}
func (users Users) List() []User {
	result := make([]User, len(users))
	i := 0
	for _, value := range users {
		result[i] = value
		i++
	}
	return result
}
func (users Users) ToString() (string, error) {
	result, err := json.Marshal(users.List())
	if err != nil {
		return "", err
	}
	return string(result), nil
}
func (users Users) Find(id string) User {
	return users[id]
}

func LoadDB(filename string) (Users, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Users{}, nil
		}
		return Users{}, err
	}

	var usersSlice []User
	json.Unmarshal(content, &usersSlice)

	users := Users{}
	for _, user := range usersSlice {
		users[user.Id] = user
	}

	return users, nil
}

func SaveDB(filename string, users Users) error {
	bytes, err := json.Marshal(users.List())
	if err != nil {
		return err
	}
	err = os.WriteFile(filename, bytes, 0644)
	if err != nil {
		return err
	}

	return nil
}

func ParseUser(args Arguments) (User, error) {
	item, ok := args["item"]
	if !ok || len(item) < 1 {
		return User{}, errors.New("-item flag has to be specified")
	}

	var user User
	err := json.Unmarshal([]byte(item), &user)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func Perform(args Arguments, writer io.Writer) error {
	filename, exists := args["fileName"]
	if !exists || len(filename) < 1 {
		return errors.New("-fileName flag has to be specified")
	}
	operation, exists := args["operation"]

	if !exists || len(operation) < 1 {
		return errors.New("-operation flag has to be specified")
	}

	users, err := LoadDB(filename)
	if err != nil {
		return err
	}

	switch operation {
	case "add":
		userToAdd, err := ParseUser(args)
		if err != nil {
			return err
		}
		err = users.Add(userToAdd)
		if err != nil {
			writer.Write([]byte(err.Error()))
		}

		SaveDB(filename, users)
	case "remove":
		userToRemove, exists := args["id"]
		if !exists || len(userToRemove) < 1 {
			return errors.New("-id flag has to be specified")
		}
		err = users.Remove(userToRemove)
		if err != nil {
			writer.Write([]byte(err.Error()))
		}
		SaveDB(filename, users)
	case "list":
		jsonString, err := users.ToString()
		if err != nil {
			return err
		}
		writer.Write([]byte(jsonString))

	case "findById":
		userToFind, exists := args["id"]
		if !exists || len(userToFind) < 1 {
			return errors.New("-id flag has to be specified")
		}
		user := users.Find(userToFind)
		jsonString, err := user.ToString()
		if err != nil {
			return err
		}
		writer.Write([]byte(jsonString))
	default:
		return fmt.Errorf("Operation %s not allowed!", operation)
	}

	return nil
}

func parseArgs() Arguments {
	result := Arguments{}
	args := os.Args[1:]
	for i := 0; i < len(args); i += 2 {
		cmd := args[i]
		value := args[i+1]
		if len(value) > 0 {
			result[cmd] = value
		}
	}

	return result
}

func main() {
	err := Perform(Arguments{
		"id":        "",
		"operation": "remove",
		"item":      "",
		"fileName":  "test.json",
	}, os.Stdout)
	if err != nil {
		panic(err)
	}
}
