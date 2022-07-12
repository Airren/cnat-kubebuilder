package controllers

import (
	"testing"
	"time"

	// "github.com/onsi/gomega"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	// "sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	// cnatv1alpha1 "www.github.com/airren/cnat-kubebuilder/api/v1alpha1"
)

func TestTimeUtilSchedule(t *testing.T) {
	t1 := "2022-07-12T14:14:05+08:00"
	d, err := timeUtilSchedule(t1)
	t.Logf("diff is %v, err: %s", d, err)

}

var c client.Client

var expectedRequest = reconcile.Request{
	NamespacedName: types.NamespacedName{
		Name:      "foo",
		Namespace: "default",
	},
}

var depKey = types.NamespacedName{Name: "foo-deployment", Namespace: "default"}

const timeout = time.Second * 5

func TestAtReconcile(t *testing.T) {
	// g := gomega.NewGomegaWithT(t)

	// instance := &cnatv1alpha1.At{
	// 	ObjectMeta: metav1.ObjectMeta{
	// 		Name:      "foo",
	// 		Namespace: "default",
	// 	},
	// }

	// // Setup the manager and the controller. Wrap the Controller Reconcile function
	// // so it writes each request to a channel when it finished.
	// mgr, err := manager.New(cfg, manager.Options{})
	// g.Expect(err).NotTo(gomega.HaveOccurred())

	// c = mgr.GetClient()

	// recFn, requests := SetupTest
}
