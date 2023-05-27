package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
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

// CreateChirp creates a new chirp and saves it to disk.
func (db *DB) CreateChirp(body string) (Chirp, error) {

	db.mux.Lock()
	defer db.mux.Unlock()

	// Read the existing data from the file
	dataRead, err := os.ReadFile(db.path)
	if err != nil {
		return Chirp{}, err
	}

	chirp := Chirp{
		Id:   0,
		Body: body,
	}

	if len(dataRead) == 0 {
		// No chirp exists yet, create a new database structure with the chirp
		chirp.Id = 1
		dbStructure := DBStructure{
			Chirps: map[int]Chirp{
				1: chirp,
			},
		}
		dataToWrite, err := json.MarshalIndent(dbStructure, "", "  ")
		if err != nil {
			return chirp, err
		}

		err = os.WriteFile(db.path, dataToWrite, 0644)
		if err != nil {
			return chirp, err
		}
	} else {
		// Chirps already exist, update the database structure
		dbStructure := DBStructure{}
		err := json.Unmarshal(dataRead, &dbStructure)
		if err != nil {
			return chirp, err
		}

		chirps := dbStructure.Chirps
		numChirps := len(chirps)
		chirp.Id = numChirps + 1

		chirps[chirp.Id] = chirp

		dataToWrite, err := json.MarshalIndent(dbStructure, "", "  ")
		if err != nil {
			return chirp, err
		}
		err = os.WriteFile(db.path, dataToWrite, 0644)
		if err != nil {
			return chirp, err
		}
	}

	return chirp, nil
}

// GetChirps returns all chirps in the database
func (db *DB) GetChirps() ([]Chirp, error) {

	db.mux.Lock()
	defer db.mux.Unlock()
	
	dataRead, err := os.ReadFile(db.path)

	if len(dataRead) == 0{
		return []Chirp{}, errors.New("no chirps present")
	}
	if err != nil {
		return []Chirp{}, err
	}
	// Chirps already exist, update the database structure
	dbStructure := DBStructure{}
	unmarshalErr := json.Unmarshal(dataRead, &dbStructure)
	if unmarshalErr != nil {
		return []Chirp{}, unmarshalErr
	}

	chirps := []Chirp{}

	for _,val := range dbStructure.Chirps{
		chirps = append(chirps,val)
		
	}

	return chirps,nil
}
