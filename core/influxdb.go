package core

import (
	"database/sql"
	"time"

	"os"

	log "github.com/ReviveNetwork/GoRevive/Log"
	client "github.com/influxdata/influxdb/client/v2"
)

// DB class to work with MySQL database
type InfluxDB struct {
	DBConnection     *sql.DB
	influxDBHost     string
	influxDBDatabase string
	influxDBUser     string
	influxDBPassword string
	batchPoints      client.BatchPoints
	client           client.Client
	batchTicker      *time.Ticker
}

func (iDB *InfluxDB) Reconnect() error {
	var err error

	iDB.client.Close()

	iDB.client, err = client.NewHTTPClient(client.HTTPConfig{
		Addr:     iDB.influxDBHost,
		Username: iDB.influxDBUser,
		Password: iDB.influxDBPassword,
	})
	if err != nil {
		return err
	}

	// Create a new point batch
	iDB.batchPoints, err = client.NewBatchPoints(client.BatchPointsConfig{
		Database:  iDB.influxDBDatabase,
		Precision: "s",
	})
	if err != nil {
		return err
	}

	return nil
}

// New will create a database connection and return the sql.DB
func (iDB *InfluxDB) New(influxDBHost string, influxDBDatabase string, influxDBUser string, influxDBPassword string) error {
	var err error

	iDB.influxDBHost = influxDBHost
	iDB.influxDBDatabase = influxDBDatabase
	iDB.influxDBUser = influxDBUser
	iDB.influxDBPassword = influxDBPassword

	iDB.client, err = client.NewHTTPClient(client.HTTPConfig{
		Addr:     iDB.influxDBHost,
		Username: iDB.influxDBUser,
		Password: iDB.influxDBPassword,
	})
	if err != nil {
		return err
	}

	// Create a new point batch
	iDB.batchPoints, err = client.NewBatchPoints(client.BatchPointsConfig{
		Database:  iDB.influxDBDatabase,
		Precision: "s",
	})
	if err != nil {
		return err
	}

	// Start regular sending every 10 seconds
	iDB.batchTicker = time.NewTicker(time.Second * 10)
	go func() {
		for range iDB.batchTicker.C {
			iDB.Flush()
		}
	}()

	return nil
}

func (iDB *InfluxDB) AddMetric(name string, tags map[string]string, fields map[string]interface{}) error {
	hostname, _ := os.Hostname()
	tags["hostname"] = hostname
	pt, err := client.NewPoint(name, tags, fields, time.Now())
	if err != nil {
		return err
	}
	iDB.batchPoints.AddPoint(pt)

	return nil
}

func (iDB *InfluxDB) Flush() {
	if err := iDB.client.Write(iDB.batchPoints); err != nil {
		log.Errorln(err)
	}

	iDB.batchPoints, _ = client.NewBatchPoints(client.BatchPointsConfig{
		Database:  iDB.influxDBDatabase,
		Precision: "s",
	})
}

func (iDB *InfluxDB) Stop() {
	iDB.batchTicker.Stop()
}
