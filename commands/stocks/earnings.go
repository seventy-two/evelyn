package stocks

var earnings = "unset"

func setEarnings(e string) {
	earnings = e
}

func getEarnings() string {
	return earnings
}
