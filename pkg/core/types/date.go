package types

import (
	"fmt"
	"time"
)

// Date is simple type to allow JSON marshal/unmarshal with format '2006-01-02'.
type Date time.Time

// MarshalJSON marshals Date to slice of bytes with date in format '2006-01-02'.
func (d *Date) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", time.Time(*d).Format("2006-01-02"))
	return []byte(stamp), nil
}

// UnmarshalJSON unmarshals a slice of bytes with a date in format '2006-01-02' to Date.
func (d *Date) UnmarshalJSON(b []byte) error {
	v, err := time.Parse("2006-01-02", string(b[1:len(b)-1]))
	if err != nil {
		return err
	}
	*d = Date(v)
	return nil
}
