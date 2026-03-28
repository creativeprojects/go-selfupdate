package selfupdate

type Repository interface {
	GetSlug() (string, string, error)
	Get() (any, error)
}
