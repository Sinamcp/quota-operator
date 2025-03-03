/*
Copyright 2021.

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

package v1alpha1

import (
	"context"
	"fmt"

	openshiftquotav1 "github.com/openshift/api/quota/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

const (
	teamLabel        = "snappcloud.io/team"
	defaultQuotaName = "default"
)

var (
	// log is for logging in this package.
	quotalog = logf.Log.WithName("quota-resource")
	C        client.Client
)

func (r *Quota) SetupWebhookWithManager(mgr ctrl.Manager) error {
	C = mgr.GetClient()
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

var _ webhook.Defaulter = &Quota{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Quota) Default() {
	quotalog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

var _ webhook.Validator = &Quota{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Quota) ValidateCreate() error {
	quotalog.Info("validate create", "name", r.Name)
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Quota) ValidateUpdate(old runtime.Object) error {
	quotalog.Info("validate update", "name", r.Name)
	ns := &corev1.Namespace{}
	err := C.Get(context.TODO(), types.NamespacedName{Name: r.GetNamespace()}, ns)
	if err != nil {
		quotalog.Error(err, "error getting namespace", "name", r.GetNamespace())
		return err
	}

	l, ok := ns.GetLabels()[teamLabel]
	if !ok {
		return fmt.Errorf("no team found for the project. please join your project to a team")
	}
	crq := &openshiftquotav1.ClusterResourceQuota{}
	err = C.Get(context.TODO(), types.NamespacedName{Name: l}, crq)
	if err != nil {
		quotalog.Error(err, "error getting clusterResourceQuota", "name", l)
		return fmt.Errorf("no team quota found. please request a quota for your team in cloud-support")
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Quota) ValidateDelete() error {
	quotalog.Info("validate delete", "name", r.Name)

	if r.GetName() != defaultQuotaName {
		return nil
	}

	ns := &corev1.Namespace{}
	err := C.Get(context.TODO(), types.NamespacedName{Name: r.GetNamespace()}, ns)
	if err != nil {
		quotalog.Error(err, "error getting namespace", "name", r.GetNamespace())
		return err
	}

	if ns.GetDeletionTimestamp() != nil {
		quotalog.Info("allowing default quota deletion as namespace is terminating", "namespace", r.GetNamespace())
		return nil
	}
	return fmt.Errorf("cannot delete default quota, you can only edit it")

}
