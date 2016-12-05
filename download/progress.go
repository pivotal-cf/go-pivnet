package download

import pb "gopkg.in/cheggaaa/pb.v1"

type Bar struct {
	*pb.ProgressBar
}

func NewBar() Bar {
	return Bar{pb.New(0)}
}

func (b Bar) SetTotal(contentLength int64) {
	b.Total = contentLength
}

func (b Bar) Kickoff() {
	b.Start()
}
