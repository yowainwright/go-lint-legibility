package p

type request struct {
	Url    string
	userId string
}

func send(req request) string {
	return req.Url
}
