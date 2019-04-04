package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/nokia/gitops-conductor/pkg/apis"
	"github.com/nokia/gitops-conductor/pkg/controller"
	"github.com/nokia/gitops-conductor/pkg/crd"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	"github.com/operator-framework/operator-sdk/pkg/leader"
	"github.com/operator-framework/operator-sdk/pkg/ready"
	sdkVersion "github.com/operator-framework/operator-sdk/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	extensionsobj "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

var log = logf.Log.WithName("cmd")

func printVersion() {
	log.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	log.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
	log.Info(fmt.Sprintf("operator-sdk Version: %v", sdkVersion.Version))
}

// ZapLoggerTo returns a new Logger implementation using Zap which logs
// to the given destination, instead of stderr.  It otherise behaves like
// ZapLogger.
func ZapLoggerTo(destWriter io.Writer, development bool) logr.Logger {
	// this basically mimics New<type>Config, but with a custom sink
	sink := zapcore.AddSync(destWriter)

	var enc zapcore.Encoder
	var lvl zap.AtomicLevel
	var opts []zap.Option
	if development {
		encCfg := zap.NewDevelopmentEncoderConfig()
		enc = zapcore.NewConsoleEncoder(encCfg)
		lvl = zap.NewAtomicLevelAt(zap.DebugLevel)
		opts = append(opts, zap.Development(), zap.AddStacktrace(zap.ErrorLevel))
	} else {
		encCfg := zap.NewProductionEncoderConfig()
		encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
		enc = zapcore.NewJSONEncoder(encCfg)
		lvl = zap.NewAtomicLevelAt(zap.InfoLevel)
		opts = append(opts, zap.AddStacktrace(zap.WarnLevel),
			zap.WrapCore(func(core zapcore.Core) zapcore.Core {
				return zapcore.NewSampler(core, time.Second, 100, 100)
			}))
	}
	opts = append(opts, zap.AddCallerSkip(1), zap.ErrorOutput(sink))
	logger := zap.New(zapcore.NewCore(&logf.KubeAwareEncoder{Encoder: enc, Verbose: development}, sink, lvl))
	logger = logger.WithOptions(opts...)
	return zapr.NewLogger(logger)
}

// ZapLogger is a Logger implementation.
// If development is true, a Zap development config will be used
// (stacktraces on warnings, no sampling), otherwise a Zap production
// config will be used (stacktraces on errors, sampling).
func ZapLogger(development bool) logr.Logger {
	return ZapLoggerTo(os.Stderr, development)
}

func main() {
	flag.Parse()

	// The logger instantiated here can be changed to any logger
	// implementing the logr.Logger interface. This logger will
	// be propagated through the whole operator, generating
	// uniform and structured logs.
	logf.SetLogger(ZapLogger(false))

	printVersion()

	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		log.Error(err, "failed to get watch namespace")
		os.Exit(1)
	}

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Become the leader before proceeding
	leader.Become(context.TODO(), "gitops-operator-lock")

	r := ready.NewFileReady()
	err = r.Set()
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}
	defer r.Unset()
	synTime := time.Duration(5 * time.Minute)
	key := os.Getenv("RECHECK_SEC")
	if key != "" {
		timeSec, err := strconv.Atoi(key)
		if err == nil {
			synTime = time.Duration(time.Duration(timeSec) * time.Second)
		}
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{Namespace: namespace,
		SyncPeriod: &synTime})
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	log.Info("Registering Components.")

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	if err := controller.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	crd.AddKnowCrds(mgr.GetScheme(), "/opt/config.yaml")

	createCrd(mgr)

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}
	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":2112", nil)

	log.Info("Starting the Cmd.")

	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "manager exited non-zero")
		os.Exit(1)
	}
}

func createCrd(mgr manager.Manager) {
	crd := getGitOpsCrd()

	client := mgr.GetClient()
	err := client.Create(context.TODO(), crd)
	if err != nil && !errors.IsAlreadyExists(err) {
		log.Error(err, "Failed to create CRD")
		os.Exit(1)
	}
}

func getGitOpsCrd() *extensionsobj.CustomResourceDefinition {
	return &extensionsobj.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apiextensions.k8s.io/v1beta1",
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "gitops.ops.dac.nokia.com",
		},
		Spec: extensionsobj.CustomResourceDefinitionSpec{
			Group: "ops.dac.nokia.com",
			Names: extensionsobj.CustomResourceDefinitionNames{
				Kind:     "GitOps",
				Singular: "gitops",
				Plural:   "gitops",
				ListKind: "GitOpsList",
			},
			Version: "v1alpha1",
			Scope:   "Namespaced",
		},
	}

}
