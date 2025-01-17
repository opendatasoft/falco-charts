package unit

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type serviceMonitorTemplateTest struct {
	suite.Suite
	chartPath   string
	releaseName string
	namespace   string
	templates   []string
}

func TestServiceMonitorTemplate(t *testing.T) {
	t.Parallel()

	chartFullPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	suite.Run(t, &serviceMonitorTemplateTest{
		Suite:       suite.Suite{},
		chartPath:   chartFullPath,
		releaseName: "k8s-metacollector-test",
		namespace:   "metacollector-test",
		templates:   []string{"templates/servicemonitor.yaml"},
	})
}

func (s *serviceMonitorTemplateTest) TestCreationDefaultValues() {
	// Render the servicemonitor and check that it has not been rendered.
	_, err := helm.RenderTemplateE(s.T(), &helm.Options{}, s.chartPath, s.releaseName, s.templates)
	s.Error(err, "should error")
	s.Equal("error while running command: exit status 1; Error: could not find template templates/servicemonitor.yaml in chart", err.Error())
}

func (s *serviceMonitorTemplateTest) TestEndpoint() {
	defaultEndpointsJSON := `[
    {
        "port": "metrics",
        "interval": "15s",
        "scrapeTimeout": "10s",
        "honorLabels": true,
        "path": "/metrics",
        "scheme": "http"
    }
]`
	var defaultEndpoints []monitoringv1.Endpoint
	err := json.Unmarshal([]byte(defaultEndpointsJSON), &defaultEndpoints)
	s.NoError(err)

	options := &helm.Options{SetValues: map[string]string{"serviceMonitor.create": "true"}}
	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)

	var svcMonitor monitoringv1.ServiceMonitor
	helm.UnmarshalK8SYaml(s.T(), output, &svcMonitor)

	s.Len(svcMonitor.Spec.Endpoints, 1, "should have only one endpoint")
	s.True(reflect.DeepEqual(svcMonitor.Spec.Endpoints[0], defaultEndpoints[0]))
}

func (s *serviceMonitorTemplateTest) TestNamespaceSelector() {
	options := &helm.Options{SetValues: map[string]string{"serviceMonitor.create": "true"}}
	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.releaseName, s.templates)

	var svcMonitor monitoringv1.ServiceMonitor
	helm.UnmarshalK8SYaml(s.T(), output, &svcMonitor)
	s.Len(svcMonitor.Spec.NamespaceSelector.MatchNames, 1)
	s.Equal("default", svcMonitor.Spec.NamespaceSelector.MatchNames[0])
}
