package controllers

import (
	"context"
	"encoding/json"
	"github.com/gabeduke/wioop/api/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/influxdata/influxdb-client-go"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"net/url"
	"path"
	"time"
)

// scrap fetches the sensor value from a wio device
func (r *WioReconciler) Scrape(wio *v1alpha1.Wio, log logr.Logger) (float64, error) {
	u, err := url.Parse(wio.Spec.BaseUrl)
	if err != nil {
		return 0, err
	}

	u.Path = path.Join(u.Path, wio.Spec.SensorID)
	u.Path = path.Join(u.Path, wio.Spec.SensorPath)

	log.V(1).Info("Constructing url", "url", u.String())

	client := &http.Client{}

	// Create a new request using http
	request, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Error(err, "unable to build request")
		return 0, err
	}

	// add authorization header to the req
	request.Header.Add("Authorization", wio.Spec.Token)

	// send it
	resp, err := client.Do(request)
	if err != nil {
		return 0, err
	}

	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	var f map[string]interface{}

	if err = json.Unmarshal([]byte(body), &f); err != nil {
		log.Error(err, "unable to unpack response")
	}

	log.V(1).Info("response", "data", f)

	value, ok := f[wio.Spec.ResponsePath].(float64)
	if !ok {
		log.Info("unable to parse value")
		return 0, nil
	}

	log.V(1).Info("read success", "value", value)
	return value, nil
}

// WriteValueToDB persists the metric to InfluxDB
func (r *WioReconciler) WriteValueToDB(wio *v1alpha1.Wio, value float64, log logr.Logger) error {
	// create new client with default option for server url authenticate by token
	influxClient := influxdb2.NewClient(influxURL, r.Config.InfluxToken)

	// user blocking write client for writes to desired bucket
	writeApi := influxClient.WriteApiBlocking(influxOrg, "Fleet IOT")

	// create point using fluent style
	p := influxdb2.NewPointWithMeasurement("fleet-metrics").
		AddTag("unit", wio.Spec.SensorPath).
		AddField(wio.Spec.SensorID, value).
		SetTime(time.Now())
	log.V(0).Info("write to influx", "name", p.Name(), "measurement", p.FieldList(), "tags", p.TagList())

	if err := writeApi.WritePoint(context.Background(), p); err != nil {
		return err
	}
	return nil
}

// UpdateStatus updates the CRD with the latest Scrape value and timestamp
func (r *WioReconciler) UpdateStatus(wio *v1alpha1.Wio, value float64, ctx context.Context, log logr.Logger) error {
	// Update status
	wio.Status.LastScrapeTime = &v1.Time{Time: time.Now()}
	wio.Status.LastScrapeValue = int(value)
	if err := r.Status().Update(ctx, wio); err != nil {
		log.Error(err, "unable to update Wio status")
		return err
	}
	return nil
}
