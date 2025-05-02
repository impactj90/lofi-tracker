//go:build windows

package afk

import (
	"syscall"
	"time"
	"unsafe"
)

type LASTINPUTINFO struct {
	CbSize uint32
	DwTime uint32
}

type winIdle struct{}

func (w *winIdle) GetIdleTime() (time.Duration, error) {
	modUser32 := syscall.NewLazyDLL("user32.dll")
	procGetLastInputInfo := modUser32.NewProc("GetLastInputInfo")

	var lii LASTINPUTINFO
	lii.CbSize = uint32(unsafe.Sizeof(lii))

	_, _, err := procGetLastInputInfo.Call(uintptr(unsafe.Pointer(&lii)))
	if err != nil && err.Error() != "The operation completed successfully." {
		return 0, err
	}

	ticks := uint32(time.Now().UnixMilli())
	idleTicks := ticks - lii.DwTime

	return time.Duration(idleTicks) * time.Millisecond, nil
}

func init() {
	idleProvider = &winIdle{}
}
