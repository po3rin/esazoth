package entity

type Task struct {
	ID                string `json:"id"`
	Completed         bool   `json:"completed"`
	StartTimeInMillis uint64 `json:"start_time_in_millis"`
}
