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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	dtoolsv1 "dev.azure.com/dsandbox/Development/_git/KuberJobController/api/v1"
)

var _ = Describe("JobRunner Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		jobrunner := &dtoolsv1.JobRunner{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind JobRunner")
			err := k8sClient.Get(ctx, typeNamespacedName, jobrunner)
			if err != nil && errors.IsNotFound(err) {
				resource := &dtoolsv1.JobRunner{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: dtoolsv1.JobRunnerSpec{
						Image:   "busybox",
						Command: []string{"sh", "-c", "echo hello"},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &dtoolsv1.JobRunner{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance JobRunner")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())

			// Cleanup the Job if it exists
			job := &batchv1.Job{}
			err = k8sClient.Get(ctx, typeNamespacedName, job)
			if err == nil {
				By("Cleanup the specific resource instance Job")
				Expect(k8sClient.Delete(ctx, job)).To(Succeed())
			}
		})

		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &JobRunnerReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			// First reconciliation: should create the Job and set status to Pending
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			// Verify Job was created
			foundJob := &batchv1.Job{}
			err = k8sClient.Get(ctx, typeNamespacedName, foundJob)
			Expect(err).NotTo(HaveOccurred())
			Expect(foundJob.Spec.Template.Spec.Containers[0].Image).To(Equal("busybox"))
			Expect(foundJob.Spec.Template.Spec.Containers[0].Command).To(Equal([]string{"sh", "-c", "echo hello"}))

			// Verify status updated to Pending
			updatedJobrunner := &dtoolsv1.JobRunner{}
			err = k8sClient.Get(ctx, typeNamespacedName, updatedJobrunner)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedJobrunner.Status.Execute).To(Equal(dtoolsv1.ExecutePending))

			// Mock the Job status to Completed (Succeeded > 0)
			foundJob.Status.Succeeded = 1
			Expect(k8sClient.Status().Update(ctx, foundJob)).To(Succeed())

			// Second reconciliation: should pick up the completed Job status
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			// Verify status updated to Completed
			err = k8sClient.Get(ctx, typeNamespacedName, updatedJobrunner)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedJobrunner.Status.Execute).To(Equal(dtoolsv1.ExecuteCompleted))
		})
	})
})
