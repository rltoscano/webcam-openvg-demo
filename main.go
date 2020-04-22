package main

import (
	"fmt"
	"log"

	"./bcmhost"
	"./egl"
	"./openvg"
	//"github.com/blackjack/webcam"
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

	var imgW, imgH int // TODO: Get width/height from image.
	// TODO: Use format that matches webcam.
	img, err := openvg.CreateImage(openvg.ImageFormatSrgbx8888, imgW, imgH, []openvg.ImageQuality{openvg.ImageQualityFaster})
	if err != nil {
		log.Fatalf("openvg: %v", err)
	}
	defer img.Destroy()

	eglDisplay.SwapBuffers(surface)

	fmt.Scanf("\n")

	// cam, err := webcam.Open("/dev/video0") // Open webcam
	// if err != nil {
	// 	panic(err.Error())
	// }
	// defer cam.Close()
	//
	// formatDesc := cam.GetSupportedFormats()
	// var formats []webcam.PixelFormat
	// for f := range formatDesc {
	// 	formats = append(formats, f)
	// }
	//
	// println("Available formats: ")
	// for i, value := range formats {
	// 	fmt.Fprintf(os.Stderr, "[%d] %s\n", i+1, formatDesc[value])
	// }
	//
	// choice := readChoice(fmt.Sprintf("Choose format [1-%d]: ", len(formats)))
	// format := formats[choice-1]
	//
	// fmt.Fprintf(os.Stderr, "Supported frame sizes for format %s\n", formatDesc[format])
	// frames := FrameSizes(cam.GetSupportedFrameSizes(format))
	// sort.Sort(frames)
	//
	// for i, value := range frames {
	// 	fmt.Fprintf(os.Stderr, "[%d] %s\n", i+1, value.GetString())
	// }
	// choice = readChoice(fmt.Sprintf("Choose format [1-%d]: ", len(frames)))
	// size := frames[choice-1]
	//
	// f, w, h, err := cam.SetImageFormat(format, uint32(size.MaxWidth), uint32(size.MaxHeight))
	//
	// if err != nil {
	// 	panic(err.Error())
	// } else {
	// 	fmt.Fprintf(os.Stderr, "Resulting image format: %s (%dx%d)\n", formatDesc[f], w, h)
	// }
	//
	// println("Press Enter to start streaming")
	// fmt.Scanf("\n")
	// err = cam.StartStreaming()
	// if err != nil {
	// 	panic(err.Error())
	// }
	//
	// var startTime time.Time
	// frameCount := 0
	// fmt.Print("\033[s")
	//
	// for {
	// 	err = cam.WaitForFrame(5 /* timeoutSeconds */)
	// 	frameCount++
	//
	// 	if startTime.IsZero() {
	// 		startTime = time.Now()
	// 	}
	//
	// 	switch err.(type) {
	// 	case nil:
	// 	case *webcam.Timeout:
	// 		fmt.Fprint(os.Stderr, err.Error())
	// 		continue
	// 	default:
	// 		panic(err.Error())
	// 	}
	//
	// 	frame, err := cam.ReadFrame()
	// 	if len(frame) != 0 {
	// 		// Process frame
	// 		fmt.Printf("\033[u\033[KFrame rate: %f", float64(frameCount)/(time.Now().Sub(startTime).Seconds()))
	// 	} else if err != nil {
	// 		panic(err.Error())
	// 	}
	// }
}

// type FrameSizes []webcam.FrameSize
//
// func (slice FrameSizes) Len() int {
// 	return len(slice)
// }
//
// //For sorting purposes
// func (slice FrameSizes) Less(i, j int) bool {
// 	ls := slice[i].MaxWidth * slice[i].MaxHeight
// 	rs := slice[j].MaxWidth * slice[j].MaxHeight
// 	return ls < rs
// }
//
// //For sorting purposes
// func (slice FrameSizes) Swap(i, j int) {
// 	slice[i], slice[j] = slice[j], slice[i]
// }
//
// func readChoice(s string) int {
// 	var i int
// 	for true {
// 		print(s)
// 		_, err := fmt.Scanf("%d\n", &i)
// 		if err != nil || i < 1 {
// 			println("Invalid input. Try again")
// 		} else {
// 			break
// 		}
// 	}
// 	return i
// }
