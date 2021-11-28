package main

import (
	"github.com/gocql/gocql"
)

type ScyllaData struct {
	session *gocql.Session
}

func NewScylla() (TodoNvm, error) {
	cluster := gocql.NewCluster("localhost")
	cluster.Keyspace = "todos"
	cluster.NumConns = 100
	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}

	createTodos := session.Query(`CREATE TABLE IF NOT EXISTS todos.todos (
		list text,
		id UUID,
		head text,
		"desc" text,
		PRIMARY KEY(list, id) );`)
	if err = createTodos.Exec(); err != nil {
		return nil, err
	}
	return &ScyllaData{session}, nil
}

func (S *ScyllaData) CreateList(list string) error {
	return nil
}

func (S *ScyllaData) RenameList(list string, name string) error {
	query := S.session.Query(`UPDATE todos.todos SET list=? WHERE list=?;`)
	return query.Bind(name, list).Exec()
}

func (S *ScyllaData) Save(list string, todo *Todo) error {
	query := S.session.Query(`INSERT INTO todos.todos (list, id, head, "desc") VALUES(?,?,?,?);`)
	id, _ := gocql.RandomUUID()
	todo.Id = id.String()
	return query.Bind(list, id, todo.Head, todo.Desc).Exec()
}

func (S *ScyllaData) Delete(list string, todo Todo) error {
	query := S.session.Query(`DELETE FROM todos.todos WHERE list = ? AND id = ?;`)
	return query.Bind(list, todo.Id).Exec()
}

func (S *ScyllaData) Get(list string) ([]Todo, error) {
	query := S.session.Query(`SELECT id, head, "desc" FROM todos.todos WHERE list=?;`).
		Bind(list)
	err := query.Exec()
	if err != nil {
		return nil, err
	}
	iter := query.Iter()
	todos := []Todo{}
	var id gocql.UUID
	var head string
	var desc string
	for iter.Scan(&id, &head, &desc) {
		todos = append(todos, Todo{id.String(), head, desc})
	}
	return todos, iter.Close()
}

func (S *ScyllaData) Update(list string, todo Todo) error {
	query := S.session.Query(`UPDATE todos.todos SET "desc"=?,head=? WHERE list=? AND id=?`)
	return query.Bind(todo.Desc, todo.Head, list, todo.Id).Exec()
}

func (S *ScyllaData) Close() error {
	S.session.Close()
	return nil
}
