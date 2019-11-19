package device

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	log "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/devices/gpu/nvidia/nvml"
	"github.com/hashicorp/nomad/helper"
	"github.com/hashicorp/nomad/helper/uuid"
	"github.com/hashicorp/nomad/plugins/shared/structs"

	"github.com/hashicorp/nomad/plugins/base"
	"github.com/hashicorp/nomad/plugins/device"
	"github.com/hashicorp/nomad/plugins/shared/hclspec"
	"github.com/kr/pretty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// pluginName is the name of the plugin
	// this is used for logging and (along with the version) for uniquely identifying
	// plugin binaries fingerprinted by the client
	pluginName = "skeleton-device"

	// plugin version allows the client to identify and use newer versions of
	// an installed plugin
	pluginVersion = "v0.1.0"

	// vendor is the label for the vendor providing the devices.
	// along with "type" and "model", this can be used when requesting devices:
	//   https://www.nomadproject.io/docs/job-specification/device.html#name
	vendor = "hashicorp"

	// deviceType is the "type" of device being returned
	deviceType = "skeleton"
)

var (
	// pluginInfo provides information used by Nomad to identify the plugin
	pluginInfo = &base.PluginInfoResponse{
		Type:              base.PluginTypeDevice,
		PluginApiVersions: []string{device.ApiVersion010},
		PluginVersion:     pluginVersion,
		Name:              pluginName,
	}

	// configSpec is the specification of the schema for this plugin's config.
	// this is used to validate the HCL for the plugin provided
	// as part of the client config:
	//   https://www.nomadproject.io/docs/configuration/plugin.html
	// options are here:
	//
	configSpec = hclspec.NewObject(map[string]*hclspec.Spec{
		"some_optional_string_with_default": hclspec.NewDefault(
			hclspec.NewAttr("", "string", false),
			hclspec.NewLiteral("\"note the escaped quotes in this literal\""),
		),
		"some_required_boolean": hclspec.NewAttr("", "bool", true),
		"some_optional_list":    hclspec.NewAttr("", "list(number)", false),
		"fingerprint_period": hclspec.NewDefault(
			hclspec.NewAttr("fingerprint_period", "string", false),
			hclspec.NewLiteral("\"1m\""),
		),
	})
)

// Config contains configuration information for the plugin.
type Config struct {
	SomeString        string `codec:"some_optional_string_with_default"`
	SomeBool          bool   `codec:"some_required_boolean"`
	SomeIntArray      []int  `codec:"some_optional_list"`
	FingerprintPeriod string `codec:"fingerprint_period"`
}

// SkeletonDevicePlugin contains a skeleton for most of the implementation of a
// device plugin.
type SkeletonDevicePlugin struct {
	logger log.Logger

	// local copies of all of the config values that we need for operation
	someString   string
	someBool     bool
	someIntArray []int

	// fingerprintPeriod the period for the fingerprinting loop
	// most plugins that fingerprint in a polling loop will
	// have
	fingerprintPeriod time.Duration

	// devices is a list of fingerprinted devices
	// most plugins will maintain, at least, a list of the devices that were
	// discovered during fingerprinting.
	devices    map[string]struct{}
	deviceLock sync.RWMutex
}

// NewPlugin returns a device plugin, used primarily by the main wrapper
//
// Plugin configuration isn't available yet, so there will typically be
// a limit to the initialization that can be performed at this point.
func NewPlugin(log log.Logger) *SkeletonDevicePlugin {
	return &SkeletonDevicePlugin{
		logger:  log.Named(pluginName),
		devices: make(map[string]struct{}),
	}
}

// PluginInfo returns information describing the plugin.
//
// This is called during Nomad client startup, while discovering and loading
// plugins.
func (d *SkeletonDevicePlugin) PluginInfo() (*base.PluginInfoResponse, error) {
	return pluginInfo, nil
}

// ConfigSchema returns the plugins configuration schema.
//
// This is called during Nomad client startup, immediately before parsing
// plugin config and calling SetConfig
func (d *SkeletonDevicePlugin) ConfigSchema() (*hclspec.Spec, error) {
	return configSpec, nil
}

// SetConfig is called by the client to pass the configuration for the plugin.
func (d *SkeletonDevicePlugin) SetConfig(c *base.Config) error {

	// decode the plugin config
	var config Config
	if err := base.MsgPackDecode(c.PluginConfig, &config); err != nil {
		return err
	}

	// save the configuration to the receiver
	// typically, we'll perform any additional validation or conversion
	// from MsgPack base types
	if config.SomeString != "some_acceptible_value" {
		return fmt.Errorf("some_optional_string_with_default was not acceptible", "value", config.SomeString)
	}
	d.someString = config.SomeString
	d.someBool = config.SomeBool
	d.someIntArray = config.SomeIntArray

	// for example, convert the poll period from a config string into a time.Duration
	period, err := time.ParseDuration(config.FingerprintPeriod)
	if err != nil {
		return fmt.Errorf("failed to parse doFingerprint period %q: %v", config.FingerprintPeriod, err)
	}
	d.fingerprintPeriod = period

	d.logger.Debug("test debug")
	d.logger.Info("config set", "config", log.Fmt("% #v", pretty.Formatter(config)))
	return nil
}

// Fingerprint streams detected devices.
// Messages should be emitted to the returned channel when there are changes
// to the devices or their health.
func (d *SkeletonDevicePlugin) Fingerprint(ctx context.Context) (<-chan *device.FingerprintResponse, error) {
	// Fingerprint returns a channel. The recommended way of organizing a plugin
	// is to pass that into a long-running goroutine and return the channel immediately.
	outCh := make(chan *device.FingerprintResponse)
	go d.doFingerprint(ctx, outCh)
	return outCh, nil
}

type SkeletonDevice struct {
	ID    string
	model string
}

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
			discoveredDevices := []SkeletonDevice{
				{
					ID:    uuid.Generate(),
					model: "modelA",
				},
				{
					ID:    uuid.Generate(),
					model: "modelB",
				},
			}

			// during fingerprinting, devices are grouped by "device group" in
			// order to facilitate scheduling
			// devices in the same device group should have the same
			// Vendor, Type, and Name ("Model")
			// Build Fingerprint response with computed groups and send it over the channel

			// Group all FingerprintDevices by DeviceName attribute
			deviceListByDeviceName := make(map[string][]*nvml.FingerprintDeviceData)
			for _, device := range fingerprintDevices {
				deviceName := device.DeviceName
				if deviceName == nil {
					// nvml driver was not able to detect device name. This kind
					// of devices are placed to single group with 'notAvailable' name
					notAvailableCopy := notAvailable
					deviceName = &notAvailableCopy
				}

				deviceListByDeviceName[*deviceName] = append(deviceListByDeviceName[*deviceName], device)
			}

			// Build Fingerprint response with computed groups and send it over the channel
			deviceGroups := make([]*device.DeviceGroup, 0, len(deviceListByDeviceName))
			for groupName, devices := range deviceListByDeviceName {
				deviceGroups = append(deviceGroups, deviceGroupFromFingerprintData(groupName, devices, commonAttributes))
			}
			devices <- device.NewFingerprint(deviceGroups...)
		}
	}
}

type reservationError struct {
	notExistingIDs []string
}

func (e *reservationError) Error() string {
	return fmt.Sprintf("unknown device IDs: %s", strings.Join(e.notExistingIDs, ","))
}

// Reserve returns information to the task driver on on how to mount the given devices.
// It may also perform any device-specific orchestration necessary to prepare the device
// for use.
func (d *SkeletonDevicePlugin) Reserve(deviceIDs []string) (*device.ContainerReservation, error) {
	if len(deviceIDs) == 0 {
		return &device.ContainerReservation{}, nil
	}

	// This pattern can be useful for some drivers to avoid a race condition where a device disappears
	// after being scheduled by the server but before the server gets an update on the fingerprint
	// channel that the device is no longer available.
	d.deviceLock.RLock()
	var notExistingIDs []string
	for _, id := range deviceIDs {
		if _, deviceIDExists := d.devices[id]; !deviceIDExists {
			notExistingIDs = append(notExistingIDs, id)
		}
	}
	d.deviceLock.RUnlock()
	if len(notExistingIDs) != 0 {
		return nil, &reservationError{notExistingIDs}
	}

	resp := &device.ContainerReservation{
		Envs:    map[string]string{},
		Mounts:  []*device.Mount{},
		Devices: []*device.DeviceSpec{},
	}

	for i, id := range deviceIDs {
		// Check if the device is known
		if _, ok := d.devices[id]; !ok {
			return nil, status.Newf(codes.InvalidArgument, "unknown device %q", id).Err()
		}

		// Mounts are used to mount host volumes into a container that may include
		// libraries, etc.
		resp.Mounts = append(resp.Mounts, &device.Mount{
			TaskPath: "/usr/lib/libsome-library.so",
			HostPath: "/usr/lib/libprobably-some-fingerprinted-or-configured-library.so",
			ReadOnly: true,
		})

		// Envs are a set of environment variables to set for the task.
		resp.Envs[fmt.Sprintf("DEVICE_%d", i)] = id

		// Devices are the set of devices to mount into the container.
		resp.Devices = append(resp.Devices, &device.DeviceSpec{
			// TaskPath is the location to mount the device in the task's file system.
			TaskPath: fmt.Sprintf("/dev/dev%d", i),
			// HostPath is the host location of the device.
			HostPath: fmt.Sprintf("/dev/devActual"),
			// CgroupPerms defines the permissions to use when mounting the device.
			CgroupPerms: "rx",
		})
	}

	return resp, nil
}

// Stats streams statistics for the detected devices.
// Messages should be emitted to the returned channel on the specified interval.
func (d *SkeletonDevicePlugin) Stats(ctx context.Context, interval time.Duration) (<-chan *device.StatsResponse, error) {
	// Similar to Fingerprint, Stats returns a channel. The recommended way of
	// organizing a plugin is to pass that into a long-running goroutine and
	// return the channel immediately.
	outCh := make(chan *device.StatsResponse)
	go d.doStats(ctx, outCh, interval)
	return outCh, nil
}

// doStats is the long running goroutine that streams device statistics
func (d *SkeletonDevicePlugin) doStats(ctx context.Context, stats chan *device.StatsResponse, interval time.Duration) {
	defer close(stats)

	// Create a timer that will fire immediately for the first detection
	ticker := time.NewTimer(0)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ticker.Reset(interval)
		}

		deviceStats, err := d.collectStats()
		if err != nil {
			stats <- &device.StatsResponse{
				Error: err,
			}
			return
		}
		if deviceStats == nil {
			continue
		}

		stats <- &device.StatsResponse{
			Groups: []*device.DeviceGroupStats{deviceStats},
		}
	}
}

func (d *SkeletonDevicePlugin) collectStats() (*device.DeviceGroupStats, error) {
	d.deviceLock.RLock()
	defer d.deviceLock.RUnlock()
	l := len(d.devices)
	if l == 0 {
		return nil, nil
	}

	now := time.Now()
	group := &device.DeviceGroupStats{
		Vendor:        vendor,
		Type:          deviceType,
		Name:          "some-model",
		InstanceStats: make(map[string]*device.DeviceStats, l),
	}

	for name, num := range d.devices {
		s := &device.DeviceStats{
			Summary: &structs.StatValue{
				IntNumeratorVal: helper.Int64ToPtr(num),
				Desc:            "Number of uses",
				Unit:            "uses",
			},
			Stats:     &structs.StatObject{},
			Timestamp: now,
		}
		group.InstanceStats[name] = s
	}

	return group, nil
}
