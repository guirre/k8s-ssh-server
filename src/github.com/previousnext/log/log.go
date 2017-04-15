package log

import (
	"fmt"

	"github.com/ausrasul/hashgen"
)

type Log struct {
	id string
}

func New() Log {
	return Log{
		id: hashgen.Get(10),
	}
}

func (l Log) Print(msg string) {
	fmt.Println(l.id, "|", msg)
}
