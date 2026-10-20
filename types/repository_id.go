package types

type RepositoryID int

// NewRepositoryID creates a repository ID from an integer
func NewRepositoryID(id int) RepositoryID {
	return RepositoryID(id)
}

func (r RepositoryID) GetSlug() (string, string, error) {
	return "", "", ErrInvalidID
}

func (r RepositoryID) Get() (any, error) {
	return int(r), nil
}

// Repository interface
var _ Repository = RepositoryID(0)
