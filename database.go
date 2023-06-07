package main

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps       map[int]Chirp `json:"chirps"`
	Users        map[int]User  `json:"users"`
	RevokeTokens []string      `json:"revoke_tokens"`
}

func NewDB(path string) (*DB, error) {
	if path == "" {
		// Use a default file name if the path is empty
		path = "database.json"
	} else {
		// Append the default file name to the provided path
		path = filepath.Join(path, "database.json")
	}

	db := DB{
		path: path,
		mux:  &sync.RWMutex{},
	}
	f, err := os.Create(db.path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return &db, nil
}

func (db *DB) CreateChirp(body string) (Chirp, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	dataRead, err := os.ReadFile(db.path)
	if err != nil {
		return Chirp{}, err
	}

	chirp := Chirp{
		Id:   0,
		Body: body,
	}

	dbStructure := DBStructure{}
	if len(dataRead) == 0 {
		chirp.Id = 1
		dbStructure.Chirps = map[int]Chirp{
			chirp.Id: chirp,
		}

	} else {
		err = json.Unmarshal(dataRead, &dbStructure)
		if err != nil {
			return Chirp{}, err
		}

		if len(dbStructure.Chirps) == 0 {
			chirp.Id = 1
			dbStructure.Chirps = map[int]Chirp{
				chirp.Id: chirp,
			}
		} else {
			chirp.Id = len(dbStructure.Chirps) + 1
			dbStructure.Chirps[chirp.Id] = chirp
		}
	}

	dataToWrite, err := json.MarshalIndent(dbStructure, "", "  ")
	if err != nil {
		return chirp, err
	}

	err = os.WriteFile(db.path, dataToWrite, 0644)
	if err != nil {
		return chirp, err
	}
	return chirp, nil
}

// GetChirps returns all chirps in the database
func (db *DB) GetChirps() ([]Chirp, error) {

	db.mux.Lock()
	defer db.mux.Unlock()

	dataRead, err := os.ReadFile(db.path)
	if err != nil {
		return []Chirp{}, err
	}

	if len(dataRead) == 0 {
		return []Chirp{}, errors.New("no chirps present")
	}

	dbStructure := DBStructure{}
	unmarshalErr := json.Unmarshal(dataRead, &dbStructure)
	if unmarshalErr != nil {
		return []Chirp{}, unmarshalErr
	}
	if len(dbStructure.Chirps) == 0 {
		return []Chirp{}, errors.New("no chirps present")
	}

	chirps := []Chirp{}

	for _, val := range dbStructure.Chirps {
		chirps = append(chirps, val)

	}

	return chirps, nil
}

func (db *DB) CreateUser(email string, password string) (User, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	// First check whether a user with this email already exists or not, if yes the return error !

	_, err := GetUser(db, email)
	// there is no error
	// the user exists !
	if err == nil {
		log.Fatal("user already exist....")
		return User{}, errors.New("user already exists")
	}

	dataRead, err := os.ReadFile(db.path)
	if err != nil {
		return User{}, err
	}

	user := User{
		Id:       0,
		Email:    email,
		Password: password,
	}

	dbStructure := DBStructure{}
	if len(dataRead) == 0 {
		user.Id = 1
		dbStructure.Users = map[int]User{
			user.Id: user,
		}
	} else {
		err = json.Unmarshal(dataRead, &dbStructure)
		if err != nil {
			return User{}, err
		}
		if len(dbStructure.Users) == 0 {
			user.Id = 1
			dbStructure.Users = map[int]User{
				user.Id: user,
			}
		} else {
			user.Id = len(dbStructure.Users) + 1
			dbStructure.Users[user.Id] = user
		}
	}

	dataToWrite, err := json.MarshalIndent(dbStructure, "", "  ")
	if err != nil {
		return user, err
	}

	err = os.WriteFile(db.path, dataToWrite, 0644)
	if err != nil {
		return user, err
	}

	return user, nil
}

func GetUser(db *DB, email string) (User, error) {
	dataRead, err := os.ReadFile(db.path)

	if err != nil {
		return User{}, err
	}
	dbStructure := DBStructure{}
	json.Unmarshal(dataRead, &dbStructure)

	users := dbStructure.Users

	for _, val := range users {
		if val.Email == email {
			return val, nil
		}
	}

	return User{}, errors.New("user not found")
}

func GetUserById(db *DB, id int) (User, error) {
	dataRead, err := os.ReadFile(db.path)

	if err != nil {
		return User{}, err
	}
	dbStructure := DBStructure{}
	json.Unmarshal(dataRead, &dbStructure)

	users := dbStructure.Users

	for _, val := range users {
		if val.Id == id {
			return val, nil
		}
	}

	return User{}, errors.New("user not found")
}

func UpdateUser(id int, newEmail string, newPassword string, db *DB) (User, error) {
	// Read the data from the database file
	dataRead, err := os.ReadFile(db.path)
	if err != nil {
		return User{}, err
	}

	// Unmarshal the data into a DBStructure variable
	var dbStructure DBStructure
	err = json.Unmarshal(dataRead, &dbStructure)
	if err != nil {
		return User{}, err
	}

	user, found := dbStructure.Users[id]

	if found {
		user.Email = newEmail
		user.Password = newPassword
		dbStructure.Users[id] = user
	} else {
		return User{}, errors.New("user not found")
	}

	// Marshal the updated dbStructure back into JSON format
	dataToWrite, err := json.MarshalIndent(dbStructure, "", "  ")
	if err != nil {
		return User{}, err
	}

	// Write the updated data back to the database file
	err = os.WriteFile(db.path, dataToWrite, 0644)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func AddRevoke(db *DB, tokenString string) error {
	dataRead, err := os.ReadFile(db.path)
	if err != nil {
		return err
	}

	dbStructure := DBStructure{}
	if len(dataRead) == 0 {
		dbStructure.RevokeTokens = []string{tokenString}
	} else {
		err = json.Unmarshal(dataRead, &dbStructure)
		if err != nil {
			log.Print("Error in unmarshalling in AddRevoke function: " + err.Error())
			return err
		}
		if dbStructure.RevokeTokens == nil {
			dbStructure.RevokeTokens = []string{tokenString}
		} else {
			dbStructure.RevokeTokens = append(dbStructure.RevokeTokens, tokenString)
		}
	}

	updatedData, err := json.MarshalIndent(dbStructure, "", "  ")
	if err != nil {
		log.Print("Error in marshalling in AddRevoke function: " + err.Error())
		return err
	}

	err = os.WriteFile(db.path, updatedData, 0644)
	if err != nil {
		return err
	}

	return nil
}

func GetRevoke(db *DB, tokenString string) bool {
	// Read the data from the database file
	dataRead, err := os.ReadFile(db.path)
	if err != nil {
		// Handle the error, e.g., log or return an error
		return false
	}

	// Unmarshal the data into a DBStructure object
	dbStructure := DBStructure{}
	if err := json.Unmarshal(dataRead, &dbStructure); err != nil {
		// Handle the error, e.g., log or return an error
		return false
	}

	// Check if the token is present in the RevokeTokens slice
	for _, revokedToken := range dbStructure.RevokeTokens {
		if revokedToken == tokenString {
			return true
		}
	}

	// Token not found in the RevokeTokens slice
	return false
}
