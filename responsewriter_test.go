package webapp

import (
	. "gopkg.in/check.v1"
	"net/http/httptest"
)

type ResponseWriterSuite struct{}

var _ = Suite(&ResponseWriterSuite{})

func (s *ResponseWriterSuite) TestResponseWriter(c *C) {
	rec := httptest.NewRecorder()
	rw := newResponseWriter(rec)

	c.Assert(rw.Written(), Equals, false)
	c.Assert(rw.Size(), Equals, 0)
	c.Assert(rw.Status(), Equals, 0)
}

func (s *ResponseWriterSuite) TestResponseWriterDataWritten(c *C) {
	rec := httptest.NewRecorder()
	rw := newResponseWriter(rec)

	rw.Write([]byte("1234"))

	c.Assert(rw.Written(), Equals, true)
	c.Assert(rw.Size(), Equals, 4)
	c.Assert(rw.Status(), Equals, 200)
}
