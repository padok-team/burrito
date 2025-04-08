# Fine-tuning the scheduling of runner pods

Burrito creates runner pods to execute plans and apply changes on your infrastructure. The scheduling of these pods can be fine-tuned to better fit your needs. (e.g. to avoid running too many pods at the same time, or to reduce the cost of your underlying infrastructure).

## Limit the number of runner pods in parallel

By default, Burrito does not limit the number of runner pods that can run in parallel. This can lead to a high number of pods running at the same time, which can be costly or can overload your infrastructure.

It is possible to limit the number of runner pods that can run in parallel by setting the `BURRITO_CONTROLLER_MAXCONCURRENTRUNNERPODS` environment variable in the controller, or by setting the `config.burrito.controller.maxConcurrentRunnerPods` value in the [Helm chart values file](https://github.com/padok-team/burrito/blob/main/deploy/charts/burrito/values.yaml).

You can also set this value in the TerraformRepository CRD by setting the `spec.maxConcurrentRunnerPods` field.

If the value of this parameter is set to `0`, there is no limit to the number of runner pods that can run in parallel.

When Burrito creates a pod, if the setting is both set in the controller and in the TerraformRepository, the TerraformRepository value will take precedence.
