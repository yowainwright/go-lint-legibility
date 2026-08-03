package p

func load(path string) (data string, err error) {
	data, err = read(path)
	if err != nil {
		return
	}

	data = normalize(data)
	err = validate(data)
	return
}
