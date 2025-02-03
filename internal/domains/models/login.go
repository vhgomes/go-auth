package models

type Login struct {
	HashedPassword string
	SessionToken   string
	CSRFToken      string
}
