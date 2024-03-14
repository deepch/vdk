package main

import (
	"github.com/deepch/vdk/format/ts"
	"log"
	"os"
)

func main() {
	f, _ := os.Open("edb9708f29b24ba9b175808d6b9df9c6541e25766d4a40209a8f903948b72f3f.ts")
	m := ts.NewDemuxer(f)
	var i int
	for {
		p, err := m.ReadPacket()
		if err != nil {
			return
		}
		if p.IsKeyFrame {
			i = 0
		}
		log.Println(i, p.Time, p.Data[4:10], len(p.Data))
		i++

	}
}
