package reporting

import (
	"context"
	"io/ioutil"
	"os/exec"
	"time"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/nokia/gitops-conductor/pkg/apis/ops/v1alpha1"
	"github.com/nokia/gitops-conductor/pkg/template"
	"github.com/nokia/gitops-conductor/plugin/proto"
	"google.golang.org/grpc"
	yaml "gopkg.in/yaml.v2"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

type updateTags struct {
	UpdateTag []*updateTag `yaml:"tags"`
}

type updateTag struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

var log = logf.Log.WithName("controller_reporter")

func SendReport(reporting *v1alpha1.Reporting, hash string, ops *v1alpha1.GitOps) error {

	conn, err := grpc.Dial(reporting.URL, grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor),
		grpc.WithStreamInterceptor(grpc_prometheus.StreamClientInterceptor))
	if err != nil {
		log.Error(err, "Failed to call service", "Service URL", reporting.URL)
		return err
	}
	defer conn.Close()

	var tags []*proto.Tags
	if reporting.Collector != "" {
		tags = collectTags(reporting.Collector, ops)
	}
	reporter := proto.NewReportClient(conn)
	_, err = reporter.GitUpdate(context.Background(), &proto.UpdateResult{
		Githash: "",
		Tags:    tags,
	})
	return err
}

func collectTags(collector string, spec *v1alpha1.GitOps) []*proto.Tags {
	var tags []*proto.Tags
	if collector != "" {

		// Create a new context and add a timeout to it
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel() // The cancel should be deferred so resources are cleaned up
		cmd := exec.CommandContext(ctx, collector)
		cmd.Dir = template.GetGitRootDir(spec)
		if spec.Spec.Templating.SourceFolder != "" {
			cmd.Dir += "/" + spec.Spec.Templating.SourceFolder
		}
		out, err := cmd.CombinedOutput()
		if ctx.Err() == context.DeadlineExceeded {
			log.Error(err, "Command timed out")
			return tags
		}
		if err != nil {
			log.Error(err, "Command failed", "output", string(out))
			return tags
		}

		//Collect the tags
		return collectTagFile()
	}

	return tags
}

func collectTagFile() []*proto.Tags {
	t := updateTags{}
	data, err := ioutil.ReadFile("/tmp/update_result.yaml")
	if err != nil {
		log.Error(err, "Failed to read tags file")
	}
	err = yaml.Unmarshal(data, &t)
	if err != nil {
		log.Error(err, "Failed to unmarshal yaml")
	}
	tags := []*proto.Tags{}
	for _, v := range t.UpdateTag {
		tags = append(tags, &proto.Tags{
			Key:   v.Key,
			Value: v.Value,
		})
	}

	return tags
}
