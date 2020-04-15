package server

type Authenticator interface {
	Authenticate(user, password string) (bool, error)
}
