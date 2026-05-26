package main

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

type DB struct {
	Master  *sql.DB
	Replica *sql.DB
}

type User struct {
	ID   int
	Name string
}

var tpl = template.Must(template.New("index").Parse(`
<!DOCTYPE html>
<html>
<head>
	<title>Read Replica Demo</title>
</head>
<body>
	<h1>Postgres Read Replica Demo</h1>
	<h1>Writes to Master</h1>
	<form method="POST" action="/create">
		<input type="text" name="name" placeholder="Enter name" />
		<button type="submit">Create User</button>
	</form>

	<h2>Users (served from replica)</h2>

	<ul>
		{{range .}}
			<li>{{.ID}} - {{.Name}}</li>
		{{end}}
	</ul>
</body>
</html>
`))

func NewDB() (*DB, error) {
	masterDSN := "postgres://postgres:postgres@localhost:5432/app?sslmode=disable"
	replicaDSN := "postgres://postgres:postgres@localhost:5433/app?sslmode=disable"

	master, err := sql.Open("postgres", masterDSN)
	if err != nil {
		return nil, err
	}

	replica, err := sql.Open("postgres", replicaDSN)
	if err != nil {
		return nil, err
	}

	if err := master.Ping(); err != nil {
		return nil, err
	}

	if err := replica.Ping(); err != nil {
		return nil, err
	}

	return &DB{
		Master:  master,
		Replica: replica,
	}, nil
}

func (db *DB) CreateUser(ctx context.Context, name string) error {
	_, err := db.Master.ExecContext(
		ctx,
		`INSERT INTO test(name) VALUES($1)`,
		name,
	)

	return err
}

func (db *DB) GetUsers(ctx context.Context) ([]User, error) {
	rows, err := db.Replica.QueryContext(
		ctx,
		`SELECT id, name FROM test ORDER BY id`,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []User

	for rows.Next() {
		var u User

		if err := rows.Scan(&u.ID, &u.Name); err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	return users, nil
}

func main() {
	db, err := NewDB()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		users, err := db.GetUsers(r.Context())
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		tpl.Execute(w, users)
	})

	http.HandleFunc("/create", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", 405)
			return
		}

		name := r.FormValue("name")

		if name == "" {
			http.Error(w, "name required", 400)
			return
		}

		err := db.CreateUser(r.Context(), name)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		fmt.Println("WRITE -> MASTER")

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	fmt.Println("server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
