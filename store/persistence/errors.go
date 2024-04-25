package persistence

type DBError struct {
	Err error
}

func NewDBError(err error) *DBError {
	return &DBError{Err: err}
}

func (e *DBError) Error() string {
	return e.Err.Error()
}
