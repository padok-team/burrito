package runner

import (
	"context"
	"crypto/sha256"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	tfjson "github.com/hashicorp/terraform-json"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	runnerutils "github.com/padok-team/burrito/internal/utils/runner"
	log "github.com/sirupsen/logrus"
)

const PlanArtifact string = "/tmp/plan.out"

// Execute the actions defined in the runner configuration. The runner must
// be initialized.
func (r *Runner) ExecAction() error {
	ann := map[string]string{}
	ref, _ := r.gitRepository.Head()
	commit := ref.Hash().String()

	switch r.config.Runner.Action {
	case "plan":
		sum, err := r.execPlan()
		if err != nil {
			return err
		}
		ann[annotations.LastPlanDate] = time.Now().Format(time.UnixDate)
		ann[annotations.LastPlanRun] = fmt.Sprintf("%s/%s", r.Run.Name, strconv.Itoa(r.Run.Status.Retries))
		ann[annotations.LastPlanSum] = sum
		ann[annotations.LastPlanCommit] = commit

	case "apply":
		sum, err := r.execApply()
		if err != nil {
			return err
		}
		ann[annotations.LastApplyDate] = time.Now().Format(time.UnixDate)
		ann[annotations.LastApplySum] = sum
		ann[annotations.LastApplyCommit] = commit
	default:
		return errors.New("unrecognized runner action, if this is happening there might be a version mismatch between the controller and runner")
	}

	err := annotations.Add(context.TODO(), r.Client, r.Layer, ann)
	if err != nil {
		log.Errorf("could not update terraform layer annotations: %s", err)
		return err
	}
	log.Infof("successfully updated terraform layer annotations")

	return nil
}

// Run the `init` command
func (r *Runner) ExecInit() error {
	log.Infof("launching terraform init in %s", r.workingDir)
	if r.exec == nil {
		err := errors.New("terraform or terragrunt binary not installed")
		return err
	}
	err := r.exec.Init(r.workingDir)
	if err != nil {
		log.Errorf("error executing terraform init: %s", err)
		return err
	}
	return nil
}

// Run the `plan` command and save the plan artifact in the datastore
// Returns the sha256 sum of the plan artifact
func (r *Runner) execPlan() (string, error) {
	log.Infof("starting terraform plan")
	if r.exec == nil {
		err := errors.New("terraform or terragrunt binary not installed")
		return "", err
	}
	err := r.exec.Plan(PlanArtifact)
	if err != nil {
		log.Errorf("error executing terraform plan: %s", err)
		return "", err
	}
	planJsonBytes, err := r.exec.Show(PlanArtifact, "json")
	if err != nil {
		log.Errorf("error getting terraform plan json: %s", err)
		return "", err
	}
	prettyPlan, err := r.exec.Show(PlanArtifact, "pretty")
	if err != nil {
		log.Errorf("error getting terraform pretty plan: %s", err)
		return "", err
	}
	log.Infof("sending plan to datastore")
	err = r.Datastore.PutPlan(r.Layer.Namespace, r.Layer.Name, r.Run.Name, strconv.Itoa(r.Run.Status.Retries), "pretty", prettyPlan)
	if err != nil {
		log.Errorf("could not put pretty plan in datastore: %s", err)
	}
	plan := &tfjson.Plan{}
	err = json.Unmarshal(planJsonBytes, plan)
	if err != nil {
		log.Errorf("error parsing terraform json plan: %s", err)
		return "", err
	}
	_, shortDiff := runnerutils.GetDiff(plan)
	err = r.Datastore.PutPlan(r.Layer.Namespace, r.Layer.Name, r.Run.Name, strconv.Itoa(r.Run.Status.Retries), "json", planJsonBytes)
	if err != nil {
		log.Errorf("could not put json plan in datastore: %s", err)
	}
	err = r.Datastore.PutPlan(r.Layer.Namespace, r.Layer.Name, r.Run.Name, strconv.Itoa(r.Run.Status.Retries), "short", []byte(shortDiff))
	if err != nil {
		log.Errorf("could not put short plan in datastore: %s", err)
	}
	planBin, err := os.ReadFile(PlanArtifact)
	if err != nil {
		log.Errorf("could not read plan output: %s", err)
		return "", err
	}
	sum := sha256.Sum256(planBin)
	err = r.Datastore.PutPlan(r.Layer.Namespace, r.Layer.Name, r.Run.Name, strconv.Itoa(r.Run.Status.Retries), "bin", planBin)
	if err != nil {
		log.Errorf("could not put plan binary in cache: %s", err)
		return "", err
	}
	log.Infof("terraform plan ran successfully")
	return b64.StdEncoding.EncodeToString(sum[:]), nil
}

// Run the `apply` command, by default with the plan artifact from the previous plan run
// Returns the sha256 sum of the plan artifact used
func (r *Runner) execApply() (string, error) {
	log.Infof("starting terraform apply")
	if r.exec == nil {
		err := errors.New("terraform or terragrunt binary not installed")
		return "", err
	}
	log.Info("getting plan binary in datastore at key")
	plan, err := r.Datastore.GetPlan(r.Layer.Namespace, r.Layer.Name, r.Run.Spec.Artifact.Run, r.Run.Spec.Artifact.Attempt, "bin")
	if err != nil {
		log.Errorf("could not get plan artifact: %s", err)
		return "", err
	}
	sum := sha256.Sum256(plan)
	err = os.WriteFile(PlanArtifact, plan, 0644)
	if err != nil {
		log.Errorf("could not write plan artifact to disk: %s", err)
		return "", err
	}
	log.Print("launching terraform apply")
	if configv1alpha1.GetApplyWithoutPlanArtifactEnabled(r.Repository, r.Layer) {
		log.Infof("applying without reusing plan artifact from previous plan run")
		err = r.exec.Apply("")
	} else {
		err = r.exec.Apply(PlanArtifact)
	}
	if err != nil {
		log.Errorf("error executing terraform apply: %s", err)
		return "", err
	}
	err = r.Datastore.PutPlan(r.Layer.Namespace, r.Layer.Name, r.Run.Name, strconv.Itoa(r.Run.Status.Retries), "short", []byte("Apply Successful"))
	if err != nil {
		log.Errorf("could not put short plan in datastore: %s", err)
	}
	log.Infof("terraform apply ran successfully")
	return b64.StdEncoding.EncodeToString(sum[:]), nil
}
