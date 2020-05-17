package main

import (
	"fmt"
	"log"
	"os"

	"../bcmhost"
	"../egl"
	"../ffmpeg"
	"../openvg"
)

func main() {
	bcmhost.Init()
	defer bcmhost.Deinit()
	w, h, err := bcmhost.GraphicsGetDisplaySize(bcmhost.DispmanxIDMainLcd)
	if err != nil {
		log.Printf("bcmhost: %v", err)
		return
	}
	fmt.Printf("Display size: %d %d\n", w, h)

	display, err := bcmhost.DispmanxDisplayOpen(bcmhost.DispmanxIDMainLcd)
	if err != nil {
		log.Printf("bcmhost: %v", err)
		return
	}
	defer display.Close()

	update, err := display.UpdateStart(0 /*priority*/)
	if err != nil {
		log.Printf("bcmhost: %v", err)
		return
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
		log.Printf("bcmhost: %v", err)
		return
	}

	eglDisplay, err := egl.GetDisplay(egl.DefaultDisplay)
	if err != nil {
		log.Printf("egl: %v", err)
		return
	}
	version, err := eglDisplay.Initialize()
	if err != nil {
		log.Printf("egl: %v", err)
		return
	}
	fmt.Printf("EGL version: %s\n", version)
	defer eglDisplay.Terminate()
	egl.BindAPI(egl.APIOpenVG)
	config, err := eglDisplay.ChooseConfig()
	if err != nil {
		log.Printf("egl: %v", err)
		return
	}
	window := bcmhost.NewDispmanxWindow(element, w, h)
	surface, err := eglDisplay.CreateWindowSurface(config, window)
	if err != nil {
		log.Printf("egl: %v", err)
		return
	}
	ctx, err := eglDisplay.CreateContext(config)
	if err != nil {
		log.Printf("egl: %v", err)
		return
	}
	err = eglDisplay.MakeCurrent(surface, ctx)
	if err != nil {
		log.Printf("egl: %v", err)
		return
	}
	// TODO(robert): Defer release the context.
	err = update.UpdateSubmit()
	if err != nil {
		log.Printf("bcmhost: %v", err)
		return
	}

	inputFormat, err := ffmpeg.NewInputFormat("v4l2")
	if err != nil {
		log.Printf("Failed to find input format")
		return
	}

	formatCtx, err := ffmpeg.NewFormatContext(os.Args[1], inputFormat)
	if err != nil {
		fmt.Printf("Failed creating format context: %v\n", err)
		return
	}
	defer formatCtx.Close()

	formatCtx.Dump()

	packet, err := formatCtx.ReadFrame()
	if err != nil {
		fmt.Printf("Failed to read frame: %v\n", err)
		return
	}

	stream, err := formatCtx.GetStream(0)
	if err != nil {
		log.Printf("Failed to get stream: %v", err)
		return
	}

	codec, err := ffmpeg.FindDecoder(stream.Codecpar().CodecID())
	if err != nil {
		log.Printf("Failed to get codec by ID: %v", err)
		return
	}

	codecCtx, err := codec.NewContext(stream.Codecpar())
	if err != nil {
		log.Printf("Failed to create new codec context: %v", err)
		return
	}
	defer codecCtx.Free()

	err = codecCtx.SendPacket(packet)
	if err != nil {
		log.Printf("Failed to send packet: %v", err)
		return
	}

	frame, err := ffmpeg.NewFrame()
	if err != nil {
		log.Printf("Failed to create new frame: %v", err)
		return
	}
	defer frame.Free()

	for {
		err = codecCtx.ReceiveFrame(&frame)
		if err == nil {
			break
		}
		fmt.Println("Couldn't get frame. Trying again...", err)
	}

	img, err := openvg.CreateImage(
		openvg.ImageFormatSrgbx8888,
		codecCtx.Width(),
		codecCtx.Height(),
		[]openvg.ImageQuality{openvg.ImageQualityNonantialiased})
	if err != nil {
		log.Printf("Failed to create image: %v", err)
		return
	}

	img.Write(
		frame.Data(),
		frame.Linesize(),
		openvg.ImageFormatSrgbx8888,
		0 /*x*/, 0, /*y*/
		codecCtx.Width(), codecCtx.Height())

	img.Draw()

	eglDisplay.SwapBuffers(surface)
	fmt.Scanln()
}
