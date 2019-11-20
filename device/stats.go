package device

import (
	"context"
	"time"

	"github.com/hashicorp/nomad/plugins/device"
	"github.com/hashicorp/nomad/plugins/shared/structs"
)

// doStats is the long running goroutine that streams device statistics
func (d *SkeletonDevicePlugin) doStats(ctx context.Context, stats chan<- *device.StatsResponse, interval time.Duration) {
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

		d.writeStatsToChannel(stats, time.Now())
	}
}

// deviceStats is what we "collect" and transform into device.DeviceStats objects.
//
// plugin implementations will likely have a native struct provided by the corresonding SDK
type deviceStats struct {
	ID         string
	deviceName string
	usedMemory int64
}

// writeStatsToChannel collects device stats, partitions devices into
// device groups, and sends the data over the provided channel.
func (d *SkeletonDevicePlugin) writeStatsToChannel(stats chan<- *device.StatsResponse, timestamp time.Time) {
	statsData, err := d.collectStats()
	if err != nil {
		d.logger.Error("failed to get device stats", "error", err)
		// Errors should returned in the Error field on the stats channel
		stats <- &device.StatsResponse{
			Error: err,
		}
		return
	}

	// group stats into device groups
	statsListByDeviceName := make(map[string][]*deviceStats)
	for _, statsItem := range statsData {
		deviceName := statsItem.deviceName
		statsListByDeviceName[deviceName] = append(statsListByDeviceName[deviceName], statsItem)
	}

	// create device.DeviceGroupStats struct for every group of stats
	deviceGroupsStats := make([]*device.DeviceGroupStats, 0, len(statsListByDeviceName))
	for groupName, groupStats := range statsListByDeviceName {
		deviceGroupsStats = append(deviceGroupsStats, statsForGroup(groupName, groupStats, timestamp))
	}

	stats <- &device.StatsResponse{
		Groups: deviceGroupsStats,
	}
}

func (d *SkeletonDevicePlugin) collectStats() ([]*deviceStats, error) {
	d.deviceLock.RLock()
	defer d.deviceLock.RUnlock()

	stats := []*deviceStats{}
	for ID, name := range d.devices {
		stats = append(stats, &deviceStats{
			ID:         ID,
			deviceName: name,
			usedMemory: 128,
		})
	}

	return stats, nil
}

// statsForGroup is a helper function that populates device.DeviceGroupStats
// for given groupName with groupStats list
func statsForGroup(groupName string, groupStats []*deviceStats, timestamp time.Time) *device.DeviceGroupStats {
	instanceStats := make(map[string]*device.DeviceStats)

	for _, statsItem := range groupStats {
		memStat := &structs.StatValue{
			IntNumeratorVal: &statsItem.usedMemory,
			Unit:            "MiB",
			Desc:            "Memory in use by the device",
		}

		instanceStats[statsItem.ID] = &device.DeviceStats{
			// Summary exposes a single summary metric that should be the most
			// informative to users.
			Summary: memStat,
			// Stats contains the verbose statistics for the device.
			Stats: &structs.StatObject{
				Attributes: map[string]*structs.StatValue{
					"Used Memory": memStat,
				},
			},
			// Timestamp is the time the statistics were collected.
			Timestamp: timestamp,
		}
	}

	return &device.DeviceGroupStats{
		Vendor:        vendor,
		Type:          deviceType,
		Name:          groupName,
		InstanceStats: instanceStats,
	}
}
