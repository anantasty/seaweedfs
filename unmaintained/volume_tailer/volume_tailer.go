package main

import (
	"flag"
	"log"
	"time"

	"github.com/joeslay/seaweedfs/weed/operation"
	"github.com/joeslay/seaweedfs/weed/security"
	"github.com/joeslay/seaweedfs/weed/storage/needle"
	util2 "github.com/joeslay/seaweedfs/weed/util"
	"github.com/spf13/viper"
	"golang.org/x/tools/godoc/util"
)

var (
	master         = flag.String("master", "localhost:9333", "master server host and port")
	volumeId       = flag.Int("volumeId", -1, "a volume id")
	rewindDuration = flag.Duration("rewind", -1, "rewind back in time. -1 means from the first entry. 0 means from now.")
	timeoutSeconds = flag.Int("timeoutSeconds", 0, "disconnect if no activity after these seconds")
	showTextFile   = flag.Bool("showTextFile", false, "display textual file content")
)

func main() {
	flag.Parse()

	util2.LoadConfiguration("security", false)
	grpcDialOption := security.LoadClientTLS(viper.Sub("grpc"), "client")

	vid := needle.VolumeId(*volumeId)

	var sinceTimeNs int64
	if *rewindDuration == 0 {
		sinceTimeNs = time.Now().UnixNano()
	} else if *rewindDuration == -1 {
		sinceTimeNs = 0
	} else if *rewindDuration > 0 {
		sinceTimeNs = time.Now().Add(-*rewindDuration).UnixNano()
	}

	err := operation.TailVolume(*master, grpcDialOption, vid, uint64(sinceTimeNs), *timeoutSeconds, func(n *needle.Needle) (err error) {
		if n.Size == 0 {
			println("-", n.String())
			return nil
		} else {
			println("+", n.String())
		}

		if *showTextFile {

			data := n.Data
			if n.IsGzipped() {
				if data, err = util2.UnGzipData(data); err != nil {
					return err
				}
			}
			if util.IsText(data) {
				println(string(data))
			}

			println("-", n.String(), "compressed", n.IsGzipped(), "original size", len(data))
		}
		return nil
	})

	if err != nil {
		log.Printf("Error VolumeTailSender volume %d: %v", vid, err)
	}

}
