package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"

	"github.com/mikkyang/id3-go/v2"

	id3 "github.com/mikkyang/id3-go"
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Fprintln(os.Stderr, "Need argument")
		return
	}

	var trackPeak, trackGain float32

	fname := os.Args[1]
	(func() {
		f, err := os.Open(fname)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		b := make([]byte, 4096)
		f.Seek(0, os.SEEK_SET)
		f.Read(b)
		i := bytes.Index(b, []byte("Info"))
		if i == -1 {
			f.Seek(-4096, os.SEEK_END)
			f.Read(b)
			i = bytes.Index(b, []byte("Info"))
		}
		if i == -1 {
			panic(fmt.Errorf("no 'Info' tag found"))
		}

		lameTag := b[i+0x78:]

		if bytes.Compare(lameTag[0:4], []byte("LAME")) != 0 {
			panic(fmt.Errorf("Info tag missing LAME tag"))
		}

		trackPeak = float32(binary.LittleEndian.Uint32(lameTag[11:15]) << 5)
		trackGain = float32(((int(lameTag[15])<<8)&0x100)|(int(lameTag[16])&0xFF)) / 10.0
		if (lameTag[15] & 0x02) == 0x02 {
			trackGain *= -1.0
		}
		fmt.Println(trackPeak)
		fmt.Println(trackGain)
	})()

	(func() {
		mp3, err := id3.Open(fname)
		if err != nil {
			panic(err)
		}
		defer mp3.Close()

		mp3.AddFrames(
			v2.NewDescTextFrame(
				v2.V23FrameTypeMap["TXXX"],
				"replaygain_track_peak",
				fmt.Sprintf("%6.2f", trackPeak),
			),
		)
		mp3.AddFrames(
			v2.NewDescTextFrame(
				v2.V23FrameTypeMap["TXXX"],
				"replaygain_track_gain",
				fmt.Sprintf("%3.2f dB", trackGain),
			),
		)
	})()
}
