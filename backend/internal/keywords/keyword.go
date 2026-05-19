package keywords

import "time"

type Keyword struct {
    ID        int64     `json:"id"`
    Value     string    `json:"value"`
    CreatedAt time.Time `json:"createdAt"`
}
