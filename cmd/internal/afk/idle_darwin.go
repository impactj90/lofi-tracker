//go:build darwin
// +build darwin

package afk

/*
#cgo LDFLAGS: -framework IOKit -framework CoreFoundation
#include <CoreFoundation/CoreFoundation.h>
#include <IOKit/IOKitLib.h>

uint64_t getIdleTime() {
	io_iterator_t iter;
	io_registry_entry_t entry = IOServiceGetMatchingService(kIOMasterPortDefault, IOServiceMatching("IOHIDSystem"));
	if (entry == 0) return 0;

	CFMutableDictionaryRef properties = 0;
	kern_return_t result = IORegistryEntryCreateCFProperties(entry, &properties, kCFAllocatorDefault, 0);
	if (result != KERN_SUCCESS || !properties) return 0;

	CFNumberRef idleTime = (CFNumberRef)CFDictionaryGetValue(properties, CFSTR("HIDIdleTime"));
	uint64_t nanoseconds = 0;
	if (idleTime) CFNumberGetValue(idleTime, kCFNumberSInt64Type, &nanoseconds);

	CFRelease(properties);
	IOObjectRelease(entry);

	return nanoseconds;
}
*/
import "C"
import "time"

type macIdle struct{}

func (m *macIdle) GetIdleTime() (time.Duration, error) {
	nanoseconds := C.getIdleTime()
	return time.Duration(nanoseconds), nil
}

func init() {
	idleProvider = &macIdle{}
}
