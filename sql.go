package main

import (
	"database/sql"
	"fmt"
	"strconv"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "123456"
	dbname   = "postgres"
)

const createTable = `CREATE TABLE IF NOT EXISTS public.todo
(
		id bigserial,
    head character varying(1024),
    "desc" character varying(1024),
    PRIMARY KEY (id)
)

TABLESPACE pg_default;

ALTER TABLE public.todo
    OWNER to postgres;`

// PostgresData NVM Todo adapter
type PostgresData struct {
	db *sql.DB
}

// NewSQL erstellt ein neuen SQL NVM (non volatile memory) Todo adapter
func NewSQL() (TodoNvm, error) {
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlconn)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(createTable)
	if err != nil {
		return nil, fmt.Errorf("create: %v", err)
	}
	return &PostgresData{db}, nil
}

func (P *PostgresData) Save(todo *Todo) error {
	var id int64
	err := P.db.QueryRow(
		`INSERT INTO public.todo (head, "desc")
		VALUES ($1, $2) RETURNING id`,
		todo.Head,
		todo.Desc).Scan(&id)
	if err != nil {
		return err
	}
	todo.Id = strconv.FormatInt(id, 10)
	return nil
}

func (P *PostgresData) Delete(todo Todo) error {
	_, err := P.db.Exec(`DELETE FROM public.todo WHERE id=$1;`, todo.Id)
	if err != nil {
		return err
	}
	return nil
}

func (P *PostgresData) Get() ([]Todo, error) {
	res, err := P.db.Query(`SELECT id,head,"desc" FROM public.todo;`)
	todos := []Todo{}
	if err != nil {
		return nil, err
	}

	var id int64
	var head string
	var desc string
	for res.Next() {
		err := res.Scan(&id, &head, &desc)
		if err != nil {
			return nil, fmt.Errorf("scan: %v", err)
		}
		todos = append(todos, Todo{strconv.FormatInt(id, 10), head, desc})
	}
	return todos, nil
}

func (P *PostgresData) Update(todo Todo) error {
	id, _ := strconv.ParseInt(todo.Id, 10, 64)
	_, err := P.db.Exec(`INSERT INTO public.todo (id, head, "desc")
		VALUES ($1, $2, $3)
		ON CONFLICT(id) DO UPDATE SET head=$2, "desc"=$3`,
		id,
		todo.Head,
		todo.Desc,
	)
	if err != nil {
		return err
	}
	return nil
}

func (P *PostgresData) Close() error {
	return P.db.Close()
}
