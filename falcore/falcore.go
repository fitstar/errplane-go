package errplane_falcore

import (
	"fmt"
	"github.com/fitstar/errplane-go"
	"github.com/fitstar/falcore"
	"net/http"
)

// Write the timings of a falcore request to errplane
type Tracer struct {
	ep *errplane.Client
}

func NewTracer(ep *errplane.Client) *Tracer {
	return &Tracer{ep: ep}
}

func (t *Tracer) Trace(req *falcore.Request, res *http.Response) {
	e := new(errplane.Event)
	e.Name = fmt.Sprintf("controllers/%v", req.Signature())
	e.Value = float64(req.EndTime.Sub(req.StartTime)) / float64(time.Milisecond)
	if t.ep != nil {
		t.ep.EnqueueEvent(e)
	} else {
		errplane.EnqueueEvent(e)
	}

	//TODO: more detail/break down by stage
}
