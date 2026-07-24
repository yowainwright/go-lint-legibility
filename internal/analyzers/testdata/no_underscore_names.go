package p

const max_retries = 3

func send_invite(user_id string) {
	retry_count := 0
	configure(user_id, retry_count)
}
