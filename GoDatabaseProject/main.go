package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/jcelliott/lumber" //file logging for go
)

const Version = "1.0.0"

type (
	Logger interface {
		Fatal(string, ...interface{})
		Error(string, ...interface{})
		Warn(string, ...interface{})
		Info(string, ...interface{})
		Debug(string, ...interface{})
		Trace(string, ...interface{})
	}

	Driver struct {
		mutex   sync.Mutex
		mutexes map[string]*sync.Mutex
		dir     string
		log     Logger
	}
)

type Options struct {
	Logger
}

func New(dir string, options *Options) (*Driver, error) {
	dir = filepath.Clean(dir)

	opts := Options{}

	if options != nil {
		opts = *options
	}

	if opts.Logger == nil {
		opts.Logger = lumber.NewConsoleLogger((lumber.INFO))
	}

	driver := Driver{
		dir:     dir,
		mutexes: make(map[string]*sync.Mutex),
		log:     opts.Logger,
	}

	//if directory already exists it will return nil
	//os.MkdirAll(dir, 0755) calls for persmission Owner: read, write, execute, Group: read, execute Others: read, execute
	if _, err := os.Stat(dir); err == nil {
		opts.Logger.Debug("Using '%s' (database already exists)\n", dir)
		return &driver, nil
	}

	opts.Logger.Debug("Creating the database at '%s'...\n", dir)
	return &driver, os.MkdirAll(dir, 0755)
}

func (d *Driver) Write(collection, resource string, v interface{}) error {
	if collection == "" {
		return fmt.Errorf("missing collection - no place to save record")
	}

	if resource == "" {
		return fmt.Errorf("missing resource - unable to save record(no name)")
	}

	mutex := d.getOrCreateMutex(collection)

	mutex.Lock()
	defer mutex.Unlock()

	dir := filepath.Join(d.dir, collection)
	fnlPath := filepath.Join(dir, resource+".json")
	tmpPath := fnlPath + ".tmp"

	if err := os.MkdirAll(dir, 0755); err != nil {
		var err error
		return err
	}

	b, err := json.MarshalIndent(v, "", "\t")

	if err != nil {
		return err
	}

	b = append(b, byte('\n'))

	if err := os.WriteFile(tmpPath, b, 0644); err != nil {
		return err
	}

	return os.Rename(tmpPath, fnlPath)
}

func (d *Driver) Read(collection, resource string, v interface{}) error {

	if collection == "" {
		return fmt.Errorf("missing collection - unable to read")
	}

	if resource == "" {
		return fmt.Errorf("missing resource - unable to read record (no name)")
	}

	record := filepath.Join(d.dir, collection, resource)

	if _, err := stat(record); err != nil {
		return err
	}

	b, err := os.ReadFile(record + ".json")
	if err != nil {
		return err
	}

	return json.Unmarshal(b, &v)
}

func (d *Driver) Readall(collection string) ([]string, error) {
	if collection == "" {
		return nil, fmt.Errorf("missing collection - unable to read")
	}

	dir := filepath.Join(d.dir, collection)

	if _, err := stat(dir); err != nil {
		return nil, err
	}

	files, _ := os.ReadDir(dir)

	var records []string

	for _, file := range files {
		b, err := os.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			return nil, err
		}

		records = append(records, string(b))
	}
	return records, nil
}

// Delete removes a resource from the collection. If the resource is a directory, it is removed recursively. If the resource is a regular file, its JSON representation is removed. In all other cases, Delete returns an error.
func (d *Driver) Delete(collection, resource string) error {

	path := filepath.Join(collection, resource)
	mutex := d.getOrCreateMutex(collection)
	mutex.Lock()
	defer mutex.Unlock()

	dir := filepath.Join(d.dir, path)

	switch fi, err := stat(dir); {
	case fi == nil, err != nil:
		return fmt.Errorf("unable to find file or directory named %v", path)

	case fi.Mode().IsDir():
		return os.RemoveAll(dir)

	case fi.Mode().IsRegular():
		return os.RemoveAll(dir + ".json")
	}
	return nil
}

func (d *Driver) getOrCreateMutex(collection string) *sync.Mutex {

	d.mutex.Lock()
	defer d.mutex.Unlock()
	m, ok := d.mutexes[collection]

	if !ok {
		m = &sync.Mutex{}
		d.mutexes[collection] = m
	}

	return m
}
func stat(path string) (fi os.FileInfo, err error) {
	if fi, err = os.Stat(path); os.IsNotExist(err) {
		fi, err = os.Stat(path + ".json")
	}
	return
}

// this Address type will allow us to store multiple types of user data
type Address struct {
	City        string
	District    string
	Subdistrict string
	Country     string
	Pincode     json.Number
}

type User struct {
	Name    string
	Age     json.Number
	Contact string
	Company string
	Address Address
	Email   string
}

func main() {
	dir := "./"

	db, err := New(dir, nil)
	if err != nil {
		fmt.Println("Error: ", err)
	}

	//EXAMPLES employees name
	employees := []User{
		{"Karn", "23", "23344333", "Myrl Tech", Address{"Bangkok", "Lat Phrao", "", "Thailand", "10230"}, "john@example.com"},
		{"Naphob", "25", "23344333", "Google", Address{"Chantaburi", "Lak Si", "", "Thailand", "10293"}, "paul@example.com"},
		{"Thanakarn", "27", "23344333", "Microsoft", Address{"Lop Buri", "Samyan", "", "Thailand", "10290"}, "robert@example.com"},
		{"Pawn", "29", "23344333", "Facebook", Address{"Tak", "Lat Yao", "", "Thailand", "22324"}, "vince@example.com"},
		{"Neo", "31", "23344333", "Remote-Teams", Address{"Trad", "Bang kaen", "", "Thailand", "10290"}, "neo@example.com"},
		{"Albert", "32", "23344333", "Dominate", Address{"Chacheongsao", "Lumlukka", "", "Thailand", "20930"}, "albert@example.com"},
	}

	for _, value := range employees {
		db.Write("users", value.Name, User{
			Name:    value.Name,
			Age:     value.Age,
			Contact: value.Contact,
			Company: value.Company,
			Address: value.Address,
		})
	}

	records, err := db.Readall("users")
	if err != nil {
		fmt.Println("Error", err)

	}
	fmt.Println(records)

	allusers := []User{}

	for _, f := range records {
		employeesFound := User{}
		if err := json.Unmarshal([]byte(f), &employeesFound); err != nil {
			fmt.Println("Error", err)
		}
		allusers = append(allusers, employeesFound)
	}

	fmt.Println((allusers))

	if err := db.Delete("users", ""); err != nil {
		fmt.Println("Error", err)
	}
}
