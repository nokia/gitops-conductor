package gitops

import (
	"context"
	"os"
	"reflect"
	"strconv"
	"time"

	opsv1alpha1 "github.com/nokia/gitops-conductor/pkg/apis/ops/v1alpha1"
	"github.com/nokia/gitops-conductor/pkg/git"
	"github.com/nokia/gitops-conductor/pkg/reporting"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	gitc "gopkg.in/src-d/go-git.v4"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_gitops")
var rensureInterval = 20 //Minutes

var (
	opsCreated = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gitops_k8s_create_total",
		Help: "The total number of create events to k8s apiserver",
	}, []string{"opscr", "branch"})
	opsUpdated = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gitops_k8s_update_total",
		Help: "The total number of update events to k8s apiserver",
	}, []string{"opscr", "branch"})
	invalidTemplates = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gitops_k8s_invalid_current",
		Help: "The total number of templates that are invalid at current state",
	},
		[]string{
			"opscr",
		},
	)
)

// Add creates a new GitOps Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	interval := os.Getenv("ENSURE_INTERVAL")
	reEnv := 20
	if interval != "" {
		i, err := strconv.Atoi(interval)
		if err != nil {
			log.Error(err, "Invalid re ensure interval")
		} else {
			reEnv = i
		}

	}
	return &ReconcileGitOps{client: mgr.GetClient(), scheme: mgr.GetScheme(), rensureInterval: reEnv}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("gitops-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource GitOps
	err = c.Watch(&source.Kind{Type: &opsv1alpha1.GitOps{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner GitOps
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &opsv1alpha1.GitOps{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileGitOps{}

// ReconcileGitOps reconciles a GitOps object
type ReconcileGitOps struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client          client.Client
	scheme          *runtime.Scheme
	rensureInterval int
}

// Reconcile reads that state of the cluster for a GitOps object and makes changes based on the state read
// and what is in the GitOps.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileGitOps) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling GitOps")

	// Fetch the GitOps instance
	instance := &opsv1alpha1.GitOps{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	if instance.Status.RootFolder == "" || !folderExist(instance.Status.RootFolder) {
		st, err := git.SetupGit(instance)
		defer r.updateStatus(instance)
		if err != nil {
			instance.Status.FailedClones += 1
			if instance.Status.FailedClones > 5 {
				log.Error(err, "Failing to clone. exit to clean up")
				os.Exit(2)
			}
			return reconcile.Result{
				RequeueAfter: (1 * time.Minute),
			}, err
		}
		if st != "" {
			instance.Status.RootFolder = st
			git.CheckoutBranch(instance)
			git.Pull(instance)
		}
	} else {
		err = git.CheckoutBranch(instance)
		if err != nil {
			if err == gitc.NoErrAlreadyUpToDate {
				//If branch is change enforce update even no git changes
				if instance.Status.Branch != instance.Spec.Branch {
					instance.Status.Branch = instance.Spec.Branch
					defer r.updateStatus(instance)
				} else {
					if !r.isOverDuration(time.Now(), instance) {
						log.Info("No git changes waiting for interval")
						return reconcile.Result{}, nil
					}
				}
			} else {
				log.Error(err, "GitPull error")
				return reconcile.Result{
					RequeueAfter: (1 * time.Minute),
				}, err
			}
		}
	}

	r.ensureDeployments(instance)
	instance.Status.Updated = time.Now().Format("15:04:05")
	defer r.updateStatus(instance)
	if instance.Spec.Reporting != nil {
		reporting.SendReport(instance.Spec.Reporting, "", instance)
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileGitOps) isOverDuration(now time.Time, instance *opsv1alpha1.GitOps) bool {
	//Repo upto date, no need for reconile objects
	if instance.Status.Updated != "" {

		lastUpdate, err := time.Parse("15:04:05", instance.Status.Updated)
		if err != nil {
			return true
		}
		curTime := now.Format("15:04:05")
		curTimeDur, err := time.Parse("15:04:05", curTime)
		dur := curTimeDur.Sub(lastUpdate)
		if dur < time.Minute*time.Duration(r.rensureInterval) {
			//Rensure deployments at least every rensureInterval even if git have not changed
			return false
		}
	}
	return true

}

func folderExist(dir string) bool {

	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

//updateStatus updates the status of the CR instance to the API server
func (r *ReconcileGitOps) updateStatus(cr *opsv1alpha1.GitOps) {
	err := r.client.Update(context.TODO(), cr)
	if err != nil {
		log.Error(err, "Failed to setup status")
	}
}

func (r *ReconcileGitOps) ensureDeployments(cr *opsv1alpha1.GitOps) error {

	templates, _, inv := git.PullTemplates(cr, "", r.scheme)
	invalidTemplates.WithLabelValues(cr.Name).Set(float64(inv))
	for _, o := range templates {

		err := r.client.Create(context.TODO(), o)
		if err != nil && !errors.IsAlreadyExists(err) {
			log.Error(err, "Failed to create")
		} else if err == nil {
			opsCreated.WithLabelValues(cr.Name, cr.Spec.Branch).Inc()
		}

		if r.filterObject(o) {
			continue
		}
		// Check if there are diffs from the object recreate then
		err = r.client.Update(context.TODO(), o)
		if err != nil && errors.IsInvalid(err) {
			switch o.(type) {
			case *corev1.Service:
				err := r.handleService(cr, o)
				if err != nil {
					return err
				}
			}
		} else if err != nil && !errors.IsAlreadyExists(err) {
			log.Error(err, "Failed to update runtime object")
		} else if err == nil {
			opsUpdated.WithLabelValues(cr.Name, cr.Spec.Branch).Inc()
		}

	}
	return nil
}

func (r *ReconcileGitOps) filterObject(o runtime.Object) bool {
	switch o.(type) {
	//Filter service accounts as temporary until using CreateOrUpdate from controllerutil
	case *corev1.ServiceAccount:
		return true
	}
	return false
}

func (r *ReconcileGitOps) handleService(cr *opsv1alpha1.GitOps, o runtime.Object) error {

	svc := &corev1.Service{}
	newSvc := o.(*corev1.Service)
	err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: newSvc.Namespace, Name: newSvc.Name}, svc)
	if err != nil && errors.IsNotFound(err) {
		return nil
	}
	//Only update if either ports or selector have changed
	if !reflect.DeepEqual(newSvc.Spec.Ports, svc.Spec.Ports) || !reflect.DeepEqual(newSvc.Spec.Selector, svc.Spec.Selector) {
		err := r.client.Delete(context.TODO(), o)
		if err != nil {
			log.Error(err, "Failed to delete current service")
			return nil
		}
		err = r.client.Create(context.TODO(), o)
		if err != nil {
			log.Error(err, "Failed to create service", "Obj", o)
			return err
		}
		opsUpdated.WithLabelValues(cr.Name, cr.Spec.Branch).Inc()
	}
	return nil
}

func (r *ReconcileGitOps) getPod(obj runtime.Object) (*corev1.PodList, error) {
	met, _ := meta.Accessor(obj)
	lab := met.GetLabels()

	plist := &corev1.PodList{}
	labelSelector := labels.SelectorFromSet(lab)
	listOpt := &client.ListOptions{Namespace: met.GetNamespace(), LabelSelector: labelSelector}
	err := r.client.List(context.TODO(), listOpt, plist)
	if err != nil {
		log.Error(err, "Failed to list pods for object", "Name", met.GetName())
	}
	return plist, err
}
