package main

import (
	"fmt"
	"log"

	_ "image/jpeg"
	_ "image/png"

	"../bcmhost"
	"../egl"
	"../openvg"
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
	eglDisplay.SwapBuffers(surface)

	fmt.Scanln()
}
