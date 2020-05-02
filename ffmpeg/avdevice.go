package ffmpeg

/*
	#cgo arm,linux pkg-config: libavdevice
	#include <libavdevice/avdevice.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func init() {
	C.avdevice_register_all()
}

// DeviceInfo wraps AvDeviceInfo
type DeviceInfo struct {
	Name        string
	Description string
}

// ListInputSources wraps avdevice_list_input_sources.
func ListInputSources(inputFormat InputFormat) ([]DeviceInfo, error) {
	var listPtr *C.AVDeviceInfoList
	result := C.avdevice_list_input_sources(inputFormat.cptr, nil, (*C.AVDictionary)(nil), &listPtr)
	if result < 0 {
		return nil, fmt.Errorf("ffmpeg: failed to list input sources: [%d] %s", result, getErrStr(result))
	}
	defer C.avdevice_free_list_devices(&listPtr)
	devices := make([]DeviceInfo, listPtr.nb_devices)
	q := uintptr(unsafe.Pointer(listPtr.devices))
	for i := 0; i < int(listPtr.nb_devices); i++ {
		device := (**C.AVDeviceInfo)(unsafe.Pointer(q))
		devices[i].Name = C.GoString((*device).device_name)
		devices[i].Description = C.GoString((*device).device_description)
		q += unsafe.Sizeof(*device)
	}
	return devices, nil
}
