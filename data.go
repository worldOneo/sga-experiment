package todo

type TodoNvm interface {
	CreateList(list string) error
	RenameList(list string, name string) error
	Save(list string, todo *Todo) error
	Delete(list string, todo Todo) error
	Get(list string) ([]Todo, error)
	Update(list string, todo Todo) error
	Close() error
}

type Todo struct {
	Id   string `json:"id,omitempty"`
	Head string `json:"head,omitempty"`
	Desc string `json:"desc,omitempty"`
}
