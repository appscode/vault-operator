/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e_test

import (
	"context"
	"fmt"
	"time"

	api "kubevault.dev/apimachinery/apis/engine/v1alpha1"
	"kubevault.dev/operator/pkg/controller"
	"kubevault.dev/operator/test/e2e/framework"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/api/rbac/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
	googleconsts "kmodules.xyz/constants/google"
)

var _ = Describe("GCP Secret Engine", func() {

	var f *framework.Invocation

	var (
		IsSecretEngineCreated = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether SecretEngine:(%s/%s) is created", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().SecretEngines(namespace).Get(context.TODO(), name, metav1.GetOptions{})
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue(), "SecretEngine is created")
		}
		IsSecretEngineDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether SecretEngine:(%s/%s) is deleted", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().SecretEngines(namespace).Get(context.TODO(), name, metav1.GetOptions{})
				return kerrors.IsNotFound(err)
			}, timeOut, pollingInterval).Should(BeTrue(), "SecretEngine is deleted")
		}
		IsSecretEngineSucceeded = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether SecretEngine:(%s/%s) is succeeded", namespace, name))
			Eventually(func() bool {
				r, err := f.CSClient.EngineV1alpha1().SecretEngines(namespace).Get(context.TODO(), name, metav1.GetOptions{})
				if err == nil {
					return r.Status.Phase == controller.SecretEnginePhaseSuccess
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "SecretEngine status is succeeded")

		}
		IsGCPRoleCreated = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether GCPRole:(%s/%s) role is created", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().GCPRoles(namespace).Get(context.TODO(), name, metav1.GetOptions{})
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue(), "GCPRole is created")
		}
		IsGCPRoleDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether GCPRole:(%s/%s) is deleted", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().GCPRoles(namespace).Get(context.TODO(), name, metav1.GetOptions{})
				return kerrors.IsNotFound(err)
			}, timeOut, pollingInterval).Should(BeTrue(), "GCPRole is deleted")
		}
		IsGCPRoleSucceeded = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether GCPRole:(%s/%s) is succeeded", namespace, name))
			Eventually(func() bool {
				r, err := f.CSClient.EngineV1alpha1().GCPRoles(namespace).Get(context.TODO(), name, metav1.GetOptions{})
				if err == nil {
					return r.Status.Phase == controller.GCPRolePhaseSuccess
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "GCPRole status is succeeded")
		}
		IsGCPRoleFailed = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether GCPRole:(%s/%s) is failed", namespace, name))
			Eventually(func() bool {
				r, err := f.CSClient.EngineV1alpha1().GCPRoles(namespace).Get(context.TODO(), name, metav1.GetOptions{})
				if err == nil {
					return r.Status.Phase != controller.GCPRolePhaseSuccess && len(r.Status.Conditions) != 0
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "GCPRole status is failed")
		}
		IsGCPAccessKeyRequestCreated = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether GCPAccessKeyRequest:(%s/%s) is created", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(namespace).Get(context.TODO(), name, metav1.GetOptions{})
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue(), "GCPAccessKeyRequest is created")
		}
		IsGCPAccessKeyRequestDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether GCPAccessKeyRequest:(%s/%s) is deleted", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(namespace).Get(context.TODO(), name, metav1.GetOptions{})
				return kerrors.IsNotFound(err)
			}, timeOut, pollingInterval).Should(BeTrue(), "GCPAccessKeyRequest is deleted")
		}
		IsGCPAKRConditionApproved = func(name, namespace string) {
			By("Checking whether GCPAccessKeyRequestConditions-> Type: Approved")
			Eventually(func() bool {
				crd, err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(namespace).Get(context.TODO(), name, metav1.GetOptions{})
				if err == nil {
					return kmapi.IsConditionTrue(crd.Status.Conditions, kmapi.ConditionRequestApproved)
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "Conditions-> Type : Approved")
		}
		IsGCPAKRConditionDenied = func(name, namespace string) {
			By("Checking whether GCPAccessKeyRequestConditions-> Type: Denied")
			Eventually(func() bool {
				crd, err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(namespace).Get(context.TODO(), name, metav1.GetOptions{})
				if err == nil {
					return kmapi.IsConditionTrue(crd.Status.Conditions, kmapi.ConditionRequestDenied)
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "Conditions-> Type: Denied")
		}
		IsGCPAccessKeySecretCreated = func(name, namespace string) {
			By("Checking whether GCPAccessKeySecret is created")
			Eventually(func() bool {
				crd, err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(namespace).Get(context.TODO(), name, metav1.GetOptions{})
				if err == nil && crd.Status.Secret != nil {
					_, err2 := f.KubeClient.CoreV1().Secrets(namespace).Get(context.TODO(), crd.Status.Secret.Name, metav1.GetOptions{})
					return err2 == nil
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "GCPAccessKeySecret is created")
		}
		IsGCPAccessKeySecretDeleted = func(secretName, namespace string) {
			By("Checking whether GCPAccessKeySecret is deleted")
			Eventually(func() bool {
				_, err2 := f.KubeClient.CoreV1().Secrets(namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
				return kerrors.IsNotFound(err2)
			}, timeOut, pollingInterval).Should(BeTrue(), "GCPAccessKeySecret is deleted")
		}
	)

	BeforeEach(func() {
		f = root.Invoke()
		if !framework.SelfHostedOperator {
			Skip("Skipping GCP secret engine tests because the operator isn't running inside cluster")
		}
	})

	AfterEach(func() {
		time.Sleep(20 * time.Second)
	})

	Describe("GCPRole", func() {

		var (
			gcpCredentials core.Secret
			gcpRole        api.GCPRole
			gcpSE          api.SecretEngine
		)

		const (
			gcpCredSecret   = "gcp-cred-3224"
			gcpRoleName     = "my-gcp-roleset-4325"
			gcpSecretEngine = "my-gcp-secretengine-3423423"
		)

		BeforeEach(func() {
			credentials := googleconsts.CredentialsFromEnv()
			if len(credentials) == 0 {
				Skip("Skipping gcp secret engine tests, empty env")
			}

			gcpCredentials = core.Secret{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      gcpCredSecret,
					Namespace: f.Namespace(),
				},
				Data: credentials,
			}
			_, err := f.KubeClient.CoreV1().Secrets(f.Namespace()).Create(context.TODO(), &gcpCredentials, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred(), "Create gcp credentials secret")

			gcpRole = api.GCPRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:      gcpRoleName,
					Namespace: f.Namespace(),
				},
				Spec: api.GCPRoleSpec{
					VaultRef: core.LocalObjectReference{
						Name: f.VaultAppRef.Name,
					},
					SecretType: "access_token",
					Project:    "appscode-ci",
					Bindings: ` resource "//cloudresourcemanager.googleapis.com/projects/appscode-ci" {
					roles = ["roles/viewer"]
				}`,
					TokenScopes: []string{"https://www.googleapis.com/auth/cloud-platform"},
				},
			}

			gcpSE = api.SecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      gcpSecretEngine,
					Namespace: f.Namespace(),
				},
				Spec: api.SecretEngineSpec{
					VaultRef: core.LocalObjectReference{
						Name: f.VaultAppRef.Name,
					},
					Path: "gcp",
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						GCP: &api.GCPConfiguration{
							CredentialSecret: gcpCredSecret,
						},
					},
				},
			}
		})

		AfterEach(func() {
			err := f.KubeClient.CoreV1().Secrets(f.Namespace()).Delete(context.TODO(), gcpCredSecret, metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred(), "Delete gcp credentials secret")
		})

		Context("Create GCPRole", func() {
			var p api.GCPRole
			var se api.SecretEngine

			BeforeEach(func() {
				p = gcpRole
				se = gcpSE
			})

			AfterEach(func() {
				By("Deleting GCPRole...")
				err := f.CSClient.EngineV1alpha1().GCPRoles(gcpRole.Namespace).Delete(context.TODO(), p.Name, metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete GCPRole")

				IsGCPRoleDeleted(p.Name, p.Namespace)

				By("Deleting SecretEngine...")
				err = f.CSClient.EngineV1alpha1().SecretEngines(se.Namespace).Delete(context.TODO(), se.Name, metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete Secret engine")

				IsSecretEngineDeleted(se.Name, se.Namespace)
			})

			It("Should be successful", func() {
				By("Creating SecretEngine...")
				_, err := f.CSClient.EngineV1alpha1().SecretEngines(se.Namespace).Create(context.TODO(), &se, metav1.CreateOptions{})
				Expect(err).NotTo(HaveOccurred(), "Create SecretEngine")

				IsSecretEngineCreated(se.Name, se.Namespace)
				IsSecretEngineSucceeded(se.Name, se.Namespace)

				By("Creating GCPRole...")
				_, err = f.CSClient.EngineV1alpha1().GCPRoles(p.Namespace).Create(context.TODO(), &p, metav1.CreateOptions{})
				Expect(err).NotTo(HaveOccurred(), "Create GCPRole")

				IsGCPRoleCreated(p.Name, p.Namespace)
				IsGCPRoleSucceeded(p.Name, p.Namespace)
			})

		})

		Context("Create GCPRole without enabling secretEngine", func() {
			var p api.GCPRole

			BeforeEach(func() {
				p = gcpRole
			})

			AfterEach(func() {
				By("Deleting GCPRole...")
				err := f.CSClient.EngineV1alpha1().GCPRoles(gcpRole.Namespace).Delete(context.TODO(), p.Name, metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete GCPRole")

				IsGCPRoleDeleted(p.Name, p.Namespace)

			})

			It("Should be failed making GCPRole", func() {

				By("Creating GCPRole...")
				_, err := f.CSClient.EngineV1alpha1().GCPRoles(p.Namespace).Create(context.TODO(), &p, metav1.CreateOptions{})
				Expect(err).NotTo(HaveOccurred(), "Create GCPRole")

				IsGCPRoleCreated(p.Name, p.Namespace)
				IsGCPRoleFailed(p.Name, p.Namespace)
			})
		})

	})

	Describe("GCPAccessKeyRequest", func() {

		var (
			gcpCredentials core.Secret
			gcpRole        api.GCPRole
			gcpSE          api.SecretEngine
			gcpAKR         api.GCPAccessKeyRequest
		)

		const (
			gcpCredSecret   = "gcp-cred-3224"
			gcpRoleName     = "my-gcp-roleset-4325"
			gcpSecretEngine = "my-gcp-secretengine-3423423"
			gcpAKRName      = "my-gcp-token-2345"
		)

		BeforeEach(func() {
			credentials := googleconsts.CredentialsFromEnv()
			if len(credentials) == 0 {
				Skip("Skipping gcp secret engine tests, empty env")
			}

			gcpCredentials = core.Secret{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      gcpCredSecret,
					Namespace: f.Namespace(),
				},
				Data: credentials,
			}
			_, err := f.KubeClient.CoreV1().Secrets(f.Namespace()).Create(context.TODO(), &gcpCredentials, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred(), "Create gcp credentials secret")

			gcpSE = api.SecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      gcpSecretEngine,
					Namespace: f.Namespace(),
				},
				Spec: api.SecretEngineSpec{
					VaultRef: core.LocalObjectReference{
						Name: f.VaultAppRef.Name,
					},
					Path: "gcp",
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						GCP: &api.GCPConfiguration{
							CredentialSecret: gcpCredSecret,
						},
					},
				},
			}
			_, err = f.CSClient.EngineV1alpha1().SecretEngines(gcpSE.Namespace).Create(context.TODO(), &gcpSE, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred(), "Create gcp SecretEngine")
			IsSecretEngineCreated(gcpSE.Name, gcpSE.Namespace)

			gcpRole = api.GCPRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:      gcpRoleName,
					Namespace: f.Namespace(),
				},
				Spec: api.GCPRoleSpec{
					VaultRef: core.LocalObjectReference{
						Name: f.VaultAppRef.Name,
					},
					SecretType: "access_token",
					Project:    "appscode-ci",
					Bindings: ` resource "//cloudresourcemanager.googleapis.com/projects/appscode-ci" {
					roles = ["roles/viewer"]
				}`,
					TokenScopes: []string{"https://www.googleapis.com/auth/cloud-platform"},
				},
			}

			gcpAKR = api.GCPAccessKeyRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      gcpAKRName,
					Namespace: f.Namespace(),
				},
				Spec: api.GCPAccessKeyRequestSpec{
					RoleRef: api.RoleRef{
						Name:      gcpRoleName,
						Namespace: f.Namespace(),
					},
					Subjects: []v1.Subject{
						{
							Kind:      "ServiceAccount",
							Name:      "sa",
							Namespace: "demo",
						},
					},
				},
			}
		})

		AfterEach(func() {
			err := f.KubeClient.CoreV1().Secrets(f.Namespace()).Delete(context.TODO(), gcpCredSecret, metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred(), "Delete gcp credentials secret")

			err = f.CSClient.EngineV1alpha1().SecretEngines(gcpSE.Namespace).Delete(context.TODO(), gcpSE.Name, metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred(), "Delete gcp SecretEngine")
			IsSecretEngineDeleted(gcpSE.Name, gcpSE.Namespace)
		})

		Context("Create, Approve, Deny GCPAccessKeyRequests", func() {
			BeforeEach(func() {
				_, err := f.CSClient.EngineV1alpha1().GCPRoles(gcpRole.Namespace).Create(context.TODO(), &gcpRole, metav1.CreateOptions{})
				Expect(err).NotTo(HaveOccurred(), "Create gcpRole")

				IsGCPRoleCreated(gcpRole.Name, gcpRole.Namespace)
				IsGCPRoleSucceeded(gcpRole.Name, gcpRole.Namespace)

			})

			AfterEach(func() {
				err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(gcpAKR.Namespace).Delete(context.TODO(), gcpAKR.Name, metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete GCPAccessKeyRequest")
				IsGCPAccessKeyRequestDeleted(gcpAKR.Name, gcpAKR.Namespace)

				err = f.CSClient.EngineV1alpha1().GCPRoles(gcpRole.Namespace).Delete(context.TODO(), gcpRole.Name, metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete gcpRole")
				IsGCPRoleDeleted(gcpRole.Name, gcpRole.Namespace)
			})

			It("Should be successful, Create GCPAccessKeyRequest", func() {
				_, err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(gcpAKR.Namespace).Create(context.TODO(), &gcpAKR, metav1.CreateOptions{})
				Expect(err).NotTo(HaveOccurred(), "Create GCPAccessKeyRequest")

				IsGCPAccessKeyRequestCreated(gcpAKR.Name, gcpAKR.Namespace)
			})

			It("Should be successful, Condition approved", func() {
				By("Creating gcpAccessKeyRequest...")
				r, err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(gcpAKR.Namespace).Create(context.TODO(), &gcpAKR, metav1.CreateOptions{})
				Expect(err).NotTo(HaveOccurred(), "Create GCPAccessKeyRequest")

				IsGCPAccessKeyRequestCreated(gcpAKR.Name, gcpAKR.Namespace)

				By("Updating GCP AccessKeyRequest status...")
				err = f.UpdateGCPAccessKeyRequestStatus(&api.GCPAccessKeyRequestStatus{
					Conditions: []kmapi.Condition{
						{
							Type:               kmapi.ConditionRequestApproved,
							Status:             core.ConditionTrue,
							LastTransitionTime: metav1.Now(),
						},
					},
					Phase: api.RequestStatusPhaseApproved,
				}, r)
				Expect(err).NotTo(HaveOccurred(), "Update conditions: Approved")
				IsGCPAKRConditionApproved(gcpAKR.Name, gcpAKR.Namespace)
			})

			It("Should be successful, Condition denied", func() {
				By("Creating gcpAccessKeyRequest...")
				r, err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(gcpAKR.Namespace).Create(context.TODO(), &gcpAKR, metav1.CreateOptions{})
				Expect(err).NotTo(HaveOccurred(), "Create GCPAccessKeyRequest")

				IsGCPAccessKeyRequestCreated(gcpAKR.Name, gcpAKR.Namespace)

				By("Updating GCP AccessKeyRequest status...")
				err = f.UpdateGCPAccessKeyRequestStatus(&api.GCPAccessKeyRequestStatus{
					Conditions: []kmapi.Condition{
						{
							Type:               kmapi.ConditionRequestDenied,
							Status:             core.ConditionTrue,
							LastTransitionTime: metav1.Now(),
						},
					},
					Phase: api.RequestStatusPhaseDenied,
				}, r)
				Expect(err).NotTo(HaveOccurred(), "Update conditions: Denied")

				IsGCPAKRConditionDenied(gcpAKR.Name, gcpAKR.Namespace)
			})
		})

		Context("Create secret where SecretType is access_token", func() {
			var (
				secretName string
			)

			BeforeEach(func() {

				By("Creating gcpRole...")
				r, err := f.CSClient.EngineV1alpha1().GCPRoles(gcpRole.Namespace).Create(context.TODO(), &gcpRole, metav1.CreateOptions{})
				Expect(err).NotTo(HaveOccurred(), "Create GCPRole")

				IsGCPRoleSucceeded(r.Name, r.Namespace)

			})

			AfterEach(func() {
				By("Deleting gcp accesskeyrequest...")
				err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(gcpAKR.Namespace).Delete(context.TODO(), gcpAKR.Name, metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete GCPAccessKeyRequest")

				IsGCPAccessKeyRequestDeleted(gcpAKR.Name, gcpAKR.Namespace)
				IsGCPAccessKeySecretDeleted(secretName, gcpAKR.Namespace)

				By("Deleting gcpRole...")
				err = f.CSClient.EngineV1alpha1().GCPRoles(gcpRole.Namespace).Delete(context.TODO(), gcpRole.Name, metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete GCPRole")

				IsGCPRoleDeleted(gcpRole.Name, gcpRole.Namespace)
			})

			It("Should be successful, Create Access Key Secret", func() {
				By("Creating gcp accessKeyRequest...")
				r, err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(gcpAKR.Namespace).Create(context.TODO(), &gcpAKR, metav1.CreateOptions{})
				Expect(err).NotTo(HaveOccurred(), "Create GCPAccessKeyRequest")

				IsGCPAccessKeyRequestCreated(gcpAKR.Name, gcpAKR.Namespace)

				By("Updating GCP AccessKeyRequest status...")
				err = f.UpdateGCPAccessKeyRequestStatus(&api.GCPAccessKeyRequestStatus{
					Conditions: []kmapi.Condition{
						{
							Type:               kmapi.ConditionRequestApproved,
							Status:             core.ConditionTrue,
							LastTransitionTime: metav1.Now(),
						},
					},
					Phase: api.RequestStatusPhaseApproved,
				}, r)

				Expect(err).NotTo(HaveOccurred(), "Update conditions: Approved")
				IsGCPAKRConditionApproved(gcpAKR.Name, gcpAKR.Namespace)

				IsGCPAccessKeySecretCreated(gcpAKR.Name, gcpAKR.Namespace)

				d, err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(gcpAKR.Namespace).Get(context.TODO(), gcpAKR.Name, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred(), "Get GCPAccessKeyRequest")
				if d.Status.Secret != nil {
					secretName = d.Status.Secret.Name
				}
			})
		})

	})
})
