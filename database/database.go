package database

type DB interface {
	GetQuestions() ([]string,error)
}

// TODO: add connection to this struct
type database struct{}

func NewDatabase() *database {
	return &database{}
}

// TODO: implement database layer
func (db *database) GetQuestions() ([]string, error) {
	return []string{"first question?", "second question?", "third question?", "fourth question?"}, nil
}
