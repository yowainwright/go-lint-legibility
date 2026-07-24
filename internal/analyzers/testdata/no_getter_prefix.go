package p

type user struct {
	owner string
}

func (u user) GetOwner() string {
	return u.owner
}
