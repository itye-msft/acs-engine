package runner

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/Azure/acs-engine/test/e2e/config"
	"github.com/Azure/acs-engine/test/e2e/metrics"
	"github.com/kelseyhightower/envconfig"
)

// Ginkgo contains all of the information needed to run the ginkgo suite of tests
type Ginkgo struct {
	GinkgoNodes string `envconfig:"GINKGO_NODES" default:"6"`
	Config      *config.Config
	Point       *metrics.Point
}

// BuildGinkgoRunner creates a new Ginkgo object
func BuildGinkgoRunner(cfg *config.Config, pt *metrics.Point) (*Ginkgo, error) {
	g := new(Ginkgo)
	if err := envconfig.Process("ginkgo", g); err != nil {
		return nil, err
	}
	g.Config = cfg
	g.Point = pt
	return g, nil
}

// Run will execute an orchestrator suite of tests
func (g *Ginkgo) Run() error {
	g.Point.SetTestStart()
	testDir := fmt.Sprintf("test/e2e/%s", g.Config.Orchestrator)
	cmd := exec.Command("ginkgo", "-slowSpecThreshold", "180", "-r", testDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		g.Point.RecordTestError()
		log.Printf("Error while trying to start ginkgo:%s\n", err)
		return err
	}

	err = cmd.Wait()
	if err != nil {
		g.Point.RecordTestError()
		if g.Config.IsKubernetes() {
			out, _ := exec.Command("kubectl", "get", "all", "--all-namespaces", "-o", "wide").CombinedOutput()
			log.Printf("Running kubectl get all:\n%s\n", out)
			out, _ = exec.Command("kubectl", "get", "nodes", "-o", "wide").CombinedOutput()
			log.Printf("Running kubectl get nodes:\n%s\n", out)
		}
		return err
	}
	g.Point.RecordTestSuccess()
	return nil
}
