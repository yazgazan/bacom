package har

// HAR represent a collection of requests
type HAR struct {
	Log struct {
		Entries []Entry
	} `json:"log"`
}

// Entry represents a HAR request/response pair
type Entry struct {
	Request  Request  `json:"request"`
	Response Response `json:"response"`
}
