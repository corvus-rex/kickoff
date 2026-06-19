package pagination

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const DefaultLimit = 500

type Request struct {
	Cursor string `form:"cursor"`
	Limit  int    `form:"limit"`
}

func (r *Request) GetLimit() int {
	if r.Limit < 1 {
		return DefaultLimit
	}
	return r.Limit
}

func EncodeCompositeCursor(createdAt time.Time, id uint) string {
	raw := fmt.Sprintf("%s|%d", createdAt.Format(time.RFC3339Nano), id)
	return base64.URLEncoding.EncodeToString([]byte(raw))
}

func DecodeCompositeCursor(cursor string) (time.Time, uint, error) {
	b, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, 0, err
	}
	parts := strings.SplitN(string(b), "|", 2)
	if len(parts) != 2 {
		return time.Time{}, 0, fmt.Errorf("invalid cursor format")
	}
	createdAt, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return time.Time{}, 0, err
	}
	id, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return time.Time{}, 0, err
	}
	return createdAt, uint(id), nil
}
