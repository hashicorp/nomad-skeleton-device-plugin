package device

import (
	"context"
	"time"

	"github.com/hashicorp/nomad/helper"
	"github.com/hashicorp/nomad/helper/uuid"
	"github.com/hashicorp/nomad/plugins/device"
	"github.com/hashicorp/nomad/plugins/shared/structs"
)

// doFingerprint is the long-running goroutine that detects device changes
func (d *SkeletonDevicePlugin) doFingerprint(ctx context.Context, devices chan *device.FingerprintResponse) {
	defer close(devices)

	// Create a timer that will fire immediately for the first detection
	ticker := time.NewTimer(0)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ticker.Reset(d.fingerprintPeriod)
		}

		d.writeFingerprintToChannel(devices)
	}
}

// fingerprintedDevice is what we "discover" and transform into device.Device objects.
//
// plugin implementations will likely have a native struct provided by the corresonding SDK
type fingerprintedDevice struct {
	ID         string
	deviceName string
	PCIBusID   string
}

// writeFingerprintToChannel collects fingerprint info, partitions devices into
// device groups, and sends the data over the provided channel.
func (d *SkeletonDevicePlugin) writeFingerprintToChannel(devices chan<- *device.FingerprintResponse) {
	// The logic for fingerprinting devices and detecting the diffs
	// will vary across devices.
	//
	// For this example, we'll create a few virtual devices on the first
	// fingerprinting.
	//
	// Subsequent loops won't do anything, and theoretically, we could just exit
	// this method. However, for non-trivial devices, fingerprinting is an on-going
	// process, useful for detecting new devices and tracking the health of
	// existing devices.
	if len(d.devices) == 0 {
		d.deviceLock.Lock()
		defer d.deviceLock.Unlock()

		// "discover" some devices
		discoveredDevices := []*fingerprintedDevice{
			{
				ID:         uuid.Generate(),
				deviceName: "modelA",
				PCIBusID:   uuid.Generate(),
			},
			{
				ID:         uuid.Generate(),
				deviceName: "modelB",
				PCIBusID:   uuid.Generate(),
			},
		}

		// during fingerprinting, devices are grouped by "device group" in
		// order to facilitate scheduling
		// devices in the same device group should have the same
		// Vendor, Type, and Name ("Model")
		// Build Fingerprint response with computed groups and send it over the channel
		deviceListByDeviceName := make(map[string][]*fingerprintedDevice)
		for _, device := range discoveredDevices {
			deviceName := device.deviceName
			deviceListByDeviceName[deviceName] = append(deviceListByDeviceName[deviceName], device)
			d.devices[device.ID] = deviceName
		}

		// Build Fingerprint response with computed groups and send it over the channel
		deviceGroups := make([]*device.DeviceGroup, 0, len(deviceListByDeviceName))
		for groupName, devices := range deviceListByDeviceName {
			deviceGroups = append(deviceGroups, deviceGroupFromFingerprintData(groupName, devices))
		}
		devices <- device.NewFingerprint(deviceGroups...)
	}
}

// deviceGroupFromFingerprintData composes deviceGroup from a slice of detected devicers
func deviceGroupFromFingerprintData(groupName string, deviceList []*fingerprintedDevice) *device.DeviceGroup {
	// deviceGroup without devices makes no sense -> return nil when no devices are provided
	if len(deviceList) == 0 {
		return nil
	}

	devices := make([]*device.Device, 0, len(deviceList))
	for _, dev := range deviceList {
		devices = append(devices, &device.Device{
			ID:      dev.ID,
			Healthy: true,
			HwLocality: &device.DeviceLocality{
				PciBusID: dev.PCIBusID,
			},
		})
	}

	deviceGroup := &device.DeviceGroup{
		Vendor:  vendor,
		Type:    deviceType,
		Name:    groupName,
		Devices: devices,
		// The device API assumes that devices with the same DeviceName have the same
		// attributes like amount of memory, power, bar1memory, etc.
		// If not, then they'll need to be split into different device groups
		// with different names.
		Attributes: map[string]*structs.Attribute{
			"attrA": {
				Int:  helper.Int64ToPtr(1024),
				Unit: "MB",
			},
			"attrB": {
				Float: helper.Float64ToPtr(10.5),
				Unit:  "MW",
			},
		},
	}
	return deviceGroup
}
