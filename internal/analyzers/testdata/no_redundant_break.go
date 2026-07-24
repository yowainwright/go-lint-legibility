package p

func label(status int) string {
	switch status {
	case 1:
		return "open"
	default:
		notify(status)
		break
	}

	return ""
}
