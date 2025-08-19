package server

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rddl-network/energy-service/internal/config"
	"github.com/rddl-network/energy-service/internal/model"
)

// initMQTT initializes the MQTT client and subscribes to the topic
func (s *Server) initMQTT() {
	cfg := config.GetConfig()
	mqttCfg := cfg.MQTT
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("ssl://%s:%d", mqttCfg.Host, mqttCfg.Port))
	opts.SetUsername(mqttCfg.Username)
	opts.SetPassword(mqttCfg.Password)
	//opts.SetClientID("")

	tlsConfig := &tls.Config{}
	opts.SetTLSConfig(tlsConfig)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Printf("MQTT connect error: %v", token.Error())
		return
	}
	s.mqttClient = client
	topic := mqttCfg.Topic
	if topic == "" {
		topic = "energy-consumption-reports"
	}
	if token := client.Subscribe(topic, 0, s.handleMQTTMessage); token.Wait() && token.Error() != nil {
		log.Printf("MQTT subscribe error: %v", token.Error())
	} else {
		log.Printf("Subscribed to MQTT topic: %s", topic)
	}
}

// handleMQTTMessage processes incoming MQTT messages as energy data
func (s *Server) handleMQTTMessage(client mqtt.Client, msg mqtt.Message) {
	var energyData model.EnergyData
	if err := json.Unmarshal(msg.Payload(), &energyData); err != nil {
		log.Printf("MQTT: Failed to decode JSON: %v", err)
		return
	}
	ctx := context.Background()
	existsPlmnt, err := s.plmntClient.IsZigbeeRegistered(energyData.ID)
	if err != nil || !existsPlmnt {
		log.Printf("MQTT: ID %s not registered in Planetmint or error: %v", energyData.ID, err)
		return
	}
	reportStatus, err := s.db.GetReportStatus(energyData.ID, energyData.Date)
	if err != nil {
		log.Printf("MQTT: Failed to check report status: %v", err)
		return
	}
	if reportStatus != "" {
		log.Printf("MQTT: report for this ID and date already exists")
		return
	}
	lastPoints, err := s.influxDBClient.GetLastPoint(ctx,
		"energy_data",
		map[string]string{
			"Inspelning": energyData.ID,
			"timezone":   energyData.TimezoneName,
		})
	if err != nil {
		log.Printf("MQTT: Failed to get last point from InfluxDB: %v", err)
		return
	}
	if lastPoints != nil && energyData.Data[0].Value < lastPoints.Fields["kW/h"].(float64) {
		log.Printf("MQTT: Incompatible data: data does not increase.")
		return
	}
	status := "valid"
	if !model.IsEnergyDataIncreasing(energyData.Data) {
		status = "invalid"
		log.Printf("MQTT: Energy data for ID %s is not increasing", energyData.ID)
	}
	err = s.db.SetReportStatus(energyData.ID, energyData.Date, status)
	if err != nil {
		log.Printf("MQTT: Failed to store report status: %v", err)
	}
	if status == "invalid" {
		log.Printf("MQTT: Energy data for ID %s is not compliant", energyData.ID)
		return
	}
	go s.writeJSON2File(energyData)
	err = s.write2InfluxDB(energyData)
	if err != nil {
		log.Printf("MQTT: Failed to write to database: %v", err)
		return
	}
	log.Printf("MQTT: Energy data received and written to database successfully")
}
