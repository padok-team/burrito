package runner

import (
	"crypto/sha256"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"strconv"

	tfjson "github.com/hashicorp/terraform-json"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	runnerutils "github.com/padok-team/burrito/internal/utils/runner"
	log "github.com/sirupsen/logrus"
)

const PlanArtifact string = "/tmp/plan.out"

func (r *Runner) plan() (string, error) {
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
	err = r.datastore.PutPlan(r.layer.Namespace, r.layer.Name, r.run.Name, strconv.Itoa(r.run.Status.Retries), "pretty", prettyPlan)
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
	err = r.datastore.PutPlan(r.layer.Namespace, r.layer.Name, r.run.Name, strconv.Itoa(r.run.Status.Retries), "json", planJsonBytes)
	if err != nil {
		log.Errorf("could not put json plan in datastore: %s", err)
	}
	err = r.datastore.PutPlan(r.layer.Namespace, r.layer.Name, r.run.Name, strconv.Itoa(r.run.Status.Retries), "short", []byte(shortDiff))
	if err != nil {
		log.Errorf("could not put short plan in datastore: %s", err)
	}
	planBin, err := os.ReadFile(PlanArtifact)
	if err != nil {
		log.Errorf("could not read plan output: %s", err)
		return "", err
	}
	sum := sha256.Sum256(planBin)
	err = r.datastore.PutPlan(r.layer.Namespace, r.layer.Name, r.run.Name, strconv.Itoa(r.run.Status.Retries), "bin", planBin)
	if err != nil {
		log.Errorf("could not put plan binary in cache: %s", err)
		return "", err
	}
	log.Infof("terraform plan ran successfully")
	return b64.StdEncoding.EncodeToString(sum[:]), nil
}

func (r *Runner) apply() (string, error) {
	log.Infof("starting terraform apply")
	if r.exec == nil {
		err := errors.New("terraform or terragrunt binary not installed")
		return "", err
	}
	log.Info("getting plan binary in datastore at key")
	plan, err := r.datastore.GetPlan(r.layer.Namespace, r.layer.Name, r.run.Spec.Artifact.Run, r.run.Spec.Artifact.Attempt, "bin")
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
	if configv1alpha1.GetApplyWithoutPlanArtifactEnabled(r.repository, r.layer) {
		log.Infof("applying without reusing plan artifact from previous plan run")
		err = r.exec.Apply("")
	} else {
		err = r.exec.Apply(PlanArtifact)
	}
	if err != nil {
		log.Errorf("error executing terraform apply: %s", err)
		return "", err
	}
	err = r.datastore.PutPlan(r.layer.Namespace, r.layer.Name, r.run.Name, strconv.Itoa(r.run.Status.Retries), "short", []byte("Apply Successful"))
	if err != nil {
		log.Errorf("could not put short plan in datastore: %s", err)
	}
	log.Infof("terraform apply ran successfully")
	return b64.StdEncoding.EncodeToString(sum[:]), nil
}
