package main

import (
	"fmt"
	"log"
	"os"
	"sort"

	_ "image/jpeg"
	_ "image/png"

	"./bcmhost"
	"./egl"
	"./openvg"
	"github.com/blackjack/webcam"
)

func main() {
	bcmhost.Init()
	defer bcmhost.Deinit()
	w, h, err := bcmhost.GraphicsGetDisplaySize(bcmhost.DispmanxIDMainLcd)
	if err != nil {
		log.Fatalf("bcmhost: %v", err)
	}
	fmt.Printf("Display size: %d %d\n", w, h)

	display, err := bcmhost.DispmanxDisplayOpen(bcmhost.DispmanxIDMainLcd)
	if err != nil {
		log.Fatalf("bcmhost: %v", err)
	}
	defer display.Close()

	update, err := display.UpdateStart(0 /*priority*/)
	if err != nil {
		log.Fatalf("bcmhost: %v", err)
	}
	dest := bcmhost.Rect{0, 0, w, h}
	src := bcmhost.Rect{0, 0, w << 16, h << 16}
	element, err := update.ElementAdd(
		1, /*layer*/
		dest,
		bcmhost.DispmanxDefaultResource,
		src,
		bcmhost.DispmanxProtectionNone)
	if err != nil {
		log.Fatalf("bcmhost: %v", err)
	}

	eglDisplay, err := egl.GetDisplay(egl.DefaultDisplay)
	if err != nil {
		log.Fatalf("egl: %v", err)
	}
	version, err := eglDisplay.Initialize()
	if err != nil {
		log.Fatalf("egl: %v", err)
	}
	fmt.Printf("EGL version: %s\n", version)
	defer eglDisplay.Terminate()
	egl.BindAPI(egl.APIOpenVG)
	config, err := eglDisplay.ChooseConfig()
	if err != nil {
		log.Fatalf("egl: %v", err)
	}
	window := bcmhost.NewDispmanxWindow(element, w, h)
	surface, err := eglDisplay.CreateWindowSurface(config, window)
	if err != nil {
		log.Fatalf("egl: %v", err)
	}
	ctx, err := eglDisplay.CreateContext(config)
	if err != nil {
		log.Fatalf("egl: %v", err)
	}
	err = eglDisplay.MakeCurrent(surface, ctx)
	if err != nil {
		log.Fatalf("egl: %v", err)
	}
	// TODO(robert): Defer release the context.
	err = update.UpdateSubmit()
	if err != nil {
		log.Fatalf("bcmhost: %v", err)
	}

	openvg.SetClearColor()
	openvg.Clear(0, 0, w, h)

	cam, err := webcam.Open("/dev/video0") // Open webcam
	if err != nil {
		panic(err.Error())
	}
	defer cam.Close()

	formatDesc := cam.GetSupportedFormats()
	var formats []webcam.PixelFormat
	for f := range formatDesc {
		formats = append(formats, f)
	}

	println("Available formats: ")
	for i, value := range formats {
		fmt.Fprintf(os.Stderr, "[%d] %s\n", i+1, formatDesc[value])
	}

	choice := readChoice(fmt.Sprintf("Choose format [1-%d]: ", len(formats)))
	format := formats[choice-1]

	fmt.Fprintf(os.Stderr, "Supported frame sizes for format %s\n", formatDesc[format])
	frames := FrameSizes(cam.GetSupportedFrameSizes(format))
	sort.Sort(frames)

	for i, value := range frames {
		fmt.Fprintf(os.Stderr, "[%d] %s\n", i+1, value.GetString())
	}
	choice = readChoice(fmt.Sprintf("Choose format [1-%d]: ", len(frames)))
	size := frames[choice-1]

	f, imgW, imgH, err := cam.SetImageFormat(format, uint32(size.MaxWidth), uint32(size.MaxHeight))

	if err != nil {
		panic(err.Error())
	} else {
		fmt.Fprintf(os.Stderr, "Resulting image format: %s (%dx%d)\n", formatDesc[f], imgW, imgH)
	}

	println("Press Enter to start streaming")
	fmt.Scanf("\n")
	err = cam.StartStreaming()
	if err != nil {
		log.Fatalf("Failed to start streaming: %v", err)
	}

	// var startTime time.Time
	// frameCount := 0
	// fmt.Print("\033[s")

	fmt.Println("Waiting for frame...")
	err = cam.WaitForFrame(5 /* timeoutSeconds */)
	if err != nil {
		log.Fatal("Timed out while waiting for webcam frame.")
	}

	// fmt.Println("Reading frame...")
	// frame, err := cam.ReadFrame()
	// if err != nil {
	// 	log.Fatalf("Failed to read frame: %v", err)
	// }
	// if len(frame) == 0 {
	// 	log.Fatal("Could not read frame; frame empty.")
	// }
	// fmt.Println("==FRAME START==")
	// for _, v := range frame {
	// 	fmt.Print(v)
	// }
	// fmt.Println("\n==FRAME END==")

	// fmt.Println("Decoding frame...")
	// cfg, _, err := image.DecodeConfig(bytes.NewReader(frame))
	// if err != nil {
	// 	log.Fatalf("image/jpeg: %v", err)
	// }
	//
	// fmt.Printf("Image config: %v\n", cfg)

	// _, _, err = image.Decode(bytes.NewReader(frame))
	// if err != nil {
	// 	log.Fatalf("image.Decode: %v", err)
	// }
}
