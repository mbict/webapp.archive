package webapp

import (
	"bytes"
	"fmt"
)

type Errors []error

func (errors Errors) String() string {
	if len(errors) == 0 {
		return ""
	}
	var buffer bytes.Buffer
	for i, err := range errors {
		text := fmt.Sprintf("Error #%02d: %s \n", (i + 1), err.Error())
		buffer.WriteString(text)
	}
	return buffer.String()
}

func (errors Errors) Error() string {
	return errors.String()
}

func (errors *Errors) Add(err ...error) {
	*errors = append(*errors, err...)
}
