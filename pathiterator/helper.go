package pathiterator

import (
	"hash/fnv"
	"time"

	"github.com/charmbracelet/log"
)

func Hash(s string) int {
	h := fnv.New32a()
	h.Write([]byte(s))
	return int(h.Sum32())
}

func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Debug("Time taken", name, elapsed)
}
