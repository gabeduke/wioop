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
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"time"

	"github.com/go-logr/logr"
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
	}

	log.V(1).Info("scrape value")
	value, err := r.Scrape(&wio, log)
	if err != nil {
		return ctrl.Result{}, err
	}

	log.V(1).Info("write to db")
	if r.WriteValueToDB(&wio, value, log); err != nil {
		return ctrl.Result{}, err
	}

	log.V(1).Info("update status")
	if r.UpdateStatus(&wio, value, ctx, log); err != nil {
		return ctrl.Result{}, err
	}

	log.V(1).Info("success")
	return ctrl.Result{RequeueAfter: requeuePeriod}, nil
}

func (r *WioReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&seeedv1alpha1.Wio{}).
		WithEventFilter(predicate.Funcs{
			DeleteFunc: func(e event.DeleteEvent) bool {
				// The reconciler adds a finalizer so we perform clean-up
				// when the delete timestamp is added
				// Suppress Delete events to avoid filtering them out in the Reconcile function
				return false
			},
		}).
		WithEventFilter(predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				oldObject := e.ObjectOld.(*seeedv1alpha1.Wio)
				newObject := e.ObjectNew.(*seeedv1alpha1.Wio)
				if oldObject.Status != newObject.Status {
					// NO enqueue request
					return false
				}
				// ENQUEUE request
				return true
			},
		}).
		Complete(r)
}
