package log

import (
	"fmt"

	"github.com/ausrasul/hashgen"
)

// Log provides users with an object for printing logging data.
type Log struct {
	id string
}

// New returns a new logging object.
func New() Log {
	return Log{
		id: hashgen.Get(10),
	}
}

// Print is used to format our id prefixed message.
func (l Log) Print(msg string) {
	fmt.Println(l.id, "|", msg)
}
