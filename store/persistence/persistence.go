package persistence

type Interface interface {
	AuthInterface
	DBExists() bool
}
