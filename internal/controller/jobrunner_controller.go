/*
Copyright 2026.

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

package controller

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	dtoolsv1 "dev.azure.com/dsandbox/Development/_git/KuberJobController/api/v1"
)

// JobRunnerReconciler reconciles a JobRunner object
type JobRunnerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=dtools.dsandbox.io,resources=jobrunners,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=dtools.dsandbox.io,resources=jobrunners/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=dtools.dsandbox.io,resources=jobrunners/finalizers,verbs=update
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *JobRunnerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Fetch the JobRunner instance
	jobRunner := &dtoolsv1.JobRunner{}
	if err := r.Get(ctx, req.NamespacedName, jobRunner); err != nil {
		if errors.IsNotFound(err) {
			log.Info("JobRunner resource not found")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get JobRunner")
		return ctrl.Result{}, err
	}

	// Check if the Job already exists, if not create a new one
	foundJob := &batchv1.Job{}
	err := r.Get(ctx, types.NamespacedName{Name: jobRunner.Name, Namespace: jobRunner.Namespace}, foundJob)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Job
		job, err := r.jobForJobRunner(jobRunner)
		if err != nil {
			log.Error(err, "Failed to define new Job resource for JobRunner")
			return ctrl.Result{}, err
		}

		if err = r.Create(ctx, job); err != nil {
			log.Error(err, "Failed to create Job", "namespace", job.Namespace, "name", job.Name)
			return ctrl.Result{}, err
		}
		log.Info("Created Job", "namespace", job.Namespace, "name", job.Name)

		// Update JobRunner status to Pending
		jobRunner.Status.Execute = dtoolsv1.ExecutePending
		if err := r.Status().Update(ctx, jobRunner); err != nil {
			log.Error(err, "Failed to update JobRunner status", "name", jobRunner.Name)
			return ctrl.Result{}, err
		}

		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Job")
		return ctrl.Result{}, err
	}

	// Map Job status to JobRunner status
	var newExecuteStatus string
	if foundJob.Status.Succeeded > 0 {
		newExecuteStatus = dtoolsv1.ExecuteCompleted
	} else if foundJob.Status.Failed > 0 {
		newExecuteStatus = dtoolsv1.ExecuteFailed
	} else if foundJob.Status.Active > 0 {
		newExecuteStatus = dtoolsv1.ExecuteRunning
	} else {
		newExecuteStatus = dtoolsv1.ExecutePending
	}

	// Update JobRunner status if it has changed
	if jobRunner.Status.Execute != newExecuteStatus {
		jobRunner.Status.Execute = newExecuteStatus
		if err := r.Status().Update(ctx, jobRunner); err != nil {
			log.Error(err, "Failed to update JobRunner status", "name", jobRunner.Name)
			return ctrl.Result{}, err
		}
		log.Info("Updated JobRunner status", "status", newExecuteStatus)
	}

	return ctrl.Result{}, nil
}

// jobForJobRunner returns a batchv1.Job object for the JobRunner custom resource
func (r *JobRunnerReconciler) jobForJobRunner(jr *dtoolsv1.JobRunner) (*batchv1.Job, error) {
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jr.Name,
			Namespace: jr.Namespace,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{{
						Name:    "runner",
						Image:   jr.Spec.Image,
						Command: jr.Spec.Command,
					}},
				},
			},
		},
	}

	// Set the owner reference on the Job so that it is garbage collected when the JobRunner is deleted
	if err := ctrl.SetControllerReference(jr, job, r.Scheme); err != nil {
		return nil, err
	}

	return job, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *JobRunnerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dtoolsv1.JobRunner{}).
		Owns(&batchv1.Job{}).
		Named("jobrunner").
		Complete(r)
}
