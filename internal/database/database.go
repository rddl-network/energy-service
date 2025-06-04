package database

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
)

// Device represents a registered device
type Device struct {
	LiquidAddress     string    `json:"liquid_address"`
	DeviceName        string    `json:"device_name"`
	DeviceType        string    `json:"device_type"`
	PlanetmintAddress string    `json:"planetmint_address"`
	Timestamp         time.Time `json:"timestamp"`
}

// Database is a LevelDB key-value store using Zigbee ID as the key
type Database struct {
	db    *leveldb.DB
	mutex sync.RWMutex
}

// NewDatabase creates a new LevelDB database
func NewDatabase() (*Database, error) {
	db, err := leveldb.OpenFile("devices.db", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	return &Database{
		db: db,
	}, nil
}

// Close closes the database
func (db *Database) Close() {
	err := db.db.Close()
	if err != nil {
		fmt.Printf("Error during database close: %v\n", err)
	}
}

// keyForZigbeeID returns the LevelDB key for a given Zigbee ID
func keyForZigbeeID(zigbeeID string) []byte {
	return []byte("device:" + zigbeeID)
}

// AddDevice adds a new device to the database
func (db *Database) AddDevice(zigbeeID, liquidAddress, deviceName, deviceType, planetmintAddress string) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	device := Device{
		LiquidAddress:     liquidAddress,
		DeviceName:        deviceName,
		DeviceType:        deviceType,
		PlanetmintAddress: planetmintAddress,
		Timestamp:         time.Now(),
	}

	// Serialize the device to JSON
	data, err := json.Marshal(device)
	if err != nil {
		return fmt.Errorf("failed to marshal device data: %v", err)
	}

	// Store in LevelDB
	err = db.db.Put(keyForZigbeeID(zigbeeID), data, nil)
	if err != nil {
		return fmt.Errorf("failed to store device: %v", err)
	}

	return nil
}

// GetDevice retrieves a device by Zigbee ID
func (db *Database) GetDevice(zigbeeID string) (Device, bool, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	var device Device

	// Get from LevelDB
	data, err := db.db.Get(keyForZigbeeID(zigbeeID), nil)
	if err == leveldb.ErrNotFound {
		return device, false, nil
	}
	if err != nil {
		return device, false, fmt.Errorf("failed to get device: %v", err)
	}

	// Deserialize the JSON data
	err = json.Unmarshal(data, &device)
	if err != nil {
		return device, false, fmt.Errorf("failed to unmarshal device data: %v", err)
	}

	return device, true, nil
}

// GetAllDevices returns all devices in the database
func (db *Database) GetAllDevices() (map[string]Device, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	devices := make(map[string]Device)

	// Iterate over all entries in LevelDB
	iter := db.db.NewIterator(nil, nil)
	defer iter.Release()

	for iter.Next() {
		key := string(iter.Key())
		if !strings.HasPrefix(key, "device:") {
			continue
		}
		zigbeeID := strings.TrimPrefix(key, "device:")
		var device Device

		// Deserialize the JSON data
		err := json.Unmarshal(iter.Value(), &device)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal device data: %v", err)
		}

		devices[zigbeeID] = device
	}

	if err := iter.Error(); err != nil {
		return nil, fmt.Errorf("iterator error: %v", err)
	}

	return devices, nil
}

// GetByLiquidAddress returns devices with a specific liquid address
func (db *Database) GetByLiquidAddress(liquidAddress string) (map[string]Device, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	result := make(map[string]Device)

	// Iterate over all entries in LevelDB
	iter := db.db.NewIterator(nil, nil)
	defer iter.Release()

	for iter.Next() {
		key := string(iter.Key())
		if !strings.HasPrefix(key, "device:") {
			continue
		}
		zigbeeID := strings.TrimPrefix(key, "device:")
		var device Device

		// Deserialize the JSON data
		err := json.Unmarshal(iter.Value(), &device)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal device data: %v", err)
		}

		if device.LiquidAddress == liquidAddress {
			result[zigbeeID] = device
		}
	}

	if err := iter.Error(); err != nil {
		return nil, fmt.Errorf("iterator error: %v", err)
	}

	return result, nil
}

// ExistsZigbeeID returns true if the Zigbee ID exists in the database
func (db *Database) ExistsZigbeeID(zigbeeID string) (bool, error) {
	_, exists, err := db.GetDevice(zigbeeID)
	return exists, err
}

// SetReportStatus stores the validation status ("valid" or "invalid") for a given ZigbeeID and date
func (db *Database) SetReportStatus(zigbeeID, date, status string) error {
	key := []byte("report:device:" + zigbeeID + ",date:" + date)
	return db.db.Put(key, []byte(status), nil)
}

// GetReportStatus retrieves the validation status for a given ZigbeeID and date
func (db *Database) GetReportStatus(zigbeeID, date string) (string, error) {
	key := []byte("report:device:" + zigbeeID + ",date:" + date)
	val, err := db.db.Get(key, nil)
	if err == leveldb.ErrNotFound {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return string(val), nil
}

// DeviceStore abstracts device DB operations for mocking
// DeviceStore is implemented by *Database and MockDatabase
// Used for dependency injection in server
//
//go:generate mockery --name=DeviceStore
type DeviceStore interface {
	GetDevice(zigbeeID string) (Device, bool, error)
	AddDevice(zigbeeID, liquidAddress, deviceName, deviceType, planetmintAddress string) error
	ExistsZigbeeID(zigbeeID string) (bool, error)
	GetAllDevices() (map[string]Device, error)
	GetByLiquidAddress(liquidAddress string) (map[string]Device, error)
	SetReportStatus(zigbeeID, date, status string) error
	GetReportStatus(zigbeeID, date string) (string, error)
}
