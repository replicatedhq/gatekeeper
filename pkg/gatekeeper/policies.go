package gatekeeper

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RunPolicies executes the status command of the CLI
func (g *Gatekeeper) RunPolicies(ctx context.Context) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.TabIndent)

	fmt.Fprintln(w, listPoliciesHeader())
	w.Flush()

	policies, err := g.GatekeeperK8sClient.PoliciesV1alpha2().AdmissionPolicies("").List(metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "list policies")
	}

	for _, policy := range policies.Items {
		columns := []string{
			policy.Spec.Name,
			"Deployed",
			policy.ObjectMeta.CreationTimestamp.Format("Mon Jan 2 15:04:05 2006"),
		}

		fmt.Fprintln(w, strings.Join(columns, "\t"))
	}
	w.Flush()

	return nil
}

func listPoliciesHeader() string {
	columns := []string{
		"Name",
		"Status",
		"Deployed At",
	}

	return strings.Join(columns, "\t")
}
