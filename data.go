package main

type TodoNvm interface {
	Save(todo *Todo) error
	Delete(todo Todo) error
	Get() ([]Todo, error)
	Update(todo Todo) error
	Close() error
}

type Todo struct {
	Id   string `json:"id,omitempty"`
	Head string `json:"head,omitempty"`
	Desc string `json:"desc,omitempty"`
}
