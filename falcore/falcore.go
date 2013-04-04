package errplane_falcore

import (
	"fmt"
	"github.com/fitstar/errplane-go"
	"github.com/fitstar/falcore"
	"net/http"
	"time"
)

// Write the timings of a falcore request to errplane
type Tracer struct {
	ep *errplane.Client
}

func NewTracer(ep *errplane.Client) *Tracer {
	return &Tracer{ep: ep}
}

type traceContextJSON struct {
	Stages []string `json:"stages"`
}

func (t *Tracer) Trace(req *falcore.Request, res *http.Response) {
	e := new(errplane.Event)
	e.Value = float64(req.EndTime.Sub(req.StartTime)) / float64(time.Millisecond)

	var lastStage *falcore.PipelineStageStat
	var context = new(traceContextJSON)
	context.Stages = make([]string, req.PipelineStageStats.Len())
	i := 0
	for e := req.PipelineStageStats.Front(); e != nil; e = e.Next() {
		pss, _ := e.Value.(*falcore.PipelineStageStat)
		context.Stages[i] = fmt.Sprintf("%v:%v", pss.Name, pss.Status)
		if pss.Type == falcore.PipelineStageTypeUpstream {
			lastStage = pss
		}
		i++
	}
	e.Context = context

	if lastStage != nil {
		e.Name = fmt.Sprintf("controllers/%v#%v", lastStage.Name, req.Signature())
	} else {
		e.Name = fmt.Sprintf("controllers/%v", req.Signature())
	}
	
	if t.ep != nil {
		t.ep.EnqueueEvent(e)
	} else {
		errplane.EnqueueEvent(e)
	}
}
