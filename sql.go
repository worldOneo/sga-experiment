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

const createTable = `
CREATE TABLE IF NOT EXISTS public.lists
(
		id bigserial,
    name character varying(1024) UNIQUE,
    PRIMARY KEY (id)
) TABLESPACE pg_default;

CREATE TABLE IF NOT EXISTS public.todo
(
		id bigserial,
		list_id bigint,
    head character varying(1024),
    "desc" character varying(1024),
    PRIMARY KEY (id),
		CONSTRAINT fk_list FOREIGN KEY(list_id) REFERENCES public.lists(id)
) TABLESPACE pg_default;

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

func (P *PostgresData) RenameList(list string, name string) error {
	_, err := P.db.Exec(
		`UPDATE public.lists SET name = $1 WHERE name = $2`,
		name,
		list)
	return err
}

func (P *PostgresData) CreateList(list string) error {
	_, err := P.db.Exec(
		`INSERT INTO public.lists (name)
		VALUES ($1)`,
		list)
	return err
}

func (P *PostgresData) Save(list string, todo *Todo) error {
	var id int64
	err := P.db.QueryRow(
		`INSERT INTO public.todo (head, "desc", list_id)
		VALUES ($1, $2, (SELECT id FROM public.lists WHERE name=$3)) RETURNING id`,
		todo.Head,
		todo.Desc,
		list).Scan(&id)
	if err != nil {
		return err
	}
	todo.Id = strconv.FormatInt(id, 10)
	return nil
}

func (P *PostgresData) Delete(list string, todo Todo) error {
	_, err := P.db.Exec(`DELETE FROM public.todo WHERE id=$1;`, todo.Id)
	if err != nil {
		return err
	}
	return nil
}

func (P *PostgresData) Get(list string) ([]Todo, error) {
	res, err := P.db.Query(`SELECT public.todo.id,head,"desc" FROM public.todo
		INNER JOIN public.lists ON public.lists.id=public.todo.list_id
		WHERE public.lists.name=$1;`, list)
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

func (P *PostgresData) Update(list string, todo Todo) error {
	id, _ := strconv.ParseInt(todo.Id, 10, 64)
	_, err := P.db.Exec(`UPDATE public.todo SET head=$2, "desc"=$3 WHERE id=$1`,
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
