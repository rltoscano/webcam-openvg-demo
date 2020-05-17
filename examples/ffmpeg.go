package main

import (
	"fmt"
	"log"

	"../ffmpeg"
)

func main() {
	inputFormat, err := ffmpeg.NewInputFormat("v4l2")
	if err != nil {
		log.Fatalf("Failed to find input format")
	}

	deviceInfos, err := ffmpeg.ListInputSources(inputFormat)
	if err != nil {
		log.Fatalf("Failed listing input sources: %v", err)
	}
	for _, info := range deviceInfos {
		fmt.Println(info.Name)
	}

	ctx, err := ffmpeg.NewFormatContext("/dev/video0", inputFormat)
	if err != nil {
		log.Fatalf("Failed creating format context: %v", err)
	}
	defer ctx.Close()

	ctx.Dump()

	pkt, err := ctx.ReadFrame()
	if err != nil {
		log.Fatalf("Failed reading frame: %v", err)
	}
	defer pkt.Free()

	fmt.Println("First byte of packet:", pkt.FirstByte())
}
