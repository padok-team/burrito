package error

type StorageError struct {
	Err error
	Nil bool
}

func (s *StorageError) Error() string {
	return s.Err.Error()
}

func NotFound(err error) bool {
	ce, ok := err.(*StorageError)
	if ok {
		return ce.Nil
	}
	return false
}

func (c *StorageError) NotFound() bool {
	return c.Nil
}
