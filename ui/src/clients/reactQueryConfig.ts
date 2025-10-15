export const reactQueryKeys = {
  layers: ['layers'],
  repositories: ['repositories'],
  layer: (namespace: string, layer: string) => ['layer', namespace, layer],
  attempts: (namespace: string, layer: string, runId: string) => [
    'run',
    namespace,
    layer,
    runId,
    'attempts'
  ],
  logs: (
    namespace: string,
    layer: string,
    runId: string,
    attemptId: number | null
  ) => ['logs', namespace, layer, runId, attemptId]
};
