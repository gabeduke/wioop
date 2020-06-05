/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"github.com/influxdata/influxdb-client-go"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	seeedv1alpha1 "github.com/gabeduke/wioop/api/v1alpha1"
)

const (
	influxURL     = "https://us-central1-1.gcp.cloud2.influxdata.com"
	influxOrg     = "c670b60f97bc7205"
	requeuePeriod = 30 * time.Second
)

// WioReconciler reconciles a Wio object
type WioReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Config Config
}

// +kubebuilder:rbac:groups=seeed.leetserve.com,resources=wios,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=seeed.leetserve.com,resources=wios/status,verbs=get;update;patch

func (r *WioReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("wio", req.NamespacedName)

	var wio seeedv1alpha1.Wio
	if err := r.Get(ctx, req.NamespacedName, &wio); err != nil {
		log.Error(err, "unable to fetch Wio")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	url := fmt.Sprintf("%s/%s/%s", wio.Spec.BaseUrl, wio.Spec.SensorID, wio.Spec.SensorPath)

	client := &http.Client{}

	// Create a new request using http
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error(err, "unable to build request")
	}

	// add authorization header to the req
	request.Header.Add("Authorization", wio.Spec.Token)

	// send it
	resp, err := client.Do(request)
	if err != nil {
		return ctrl.Result{}, err
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
		return ctrl.Result{RequeueAfter: requeuePeriod}, nil
	}

	log.V(1).Info("read success", "value", value)

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
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

func (r *WioReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&seeedv1alpha1.Wio{}).
		Complete(r)
}
