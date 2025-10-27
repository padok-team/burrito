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
  plan: (
    namespace: string,
    layer: string,
    runId: string | null,
    attemptId: number | null
  ) => ['plan', namespace, layer, runId ?? 'none', attemptId ?? 'none'],
  logs: (
    namespace: string,
    layer: string,
    runId: string,
    attemptId: number | null
  ) => ['logs', namespace, layer, runId, attemptId]
};
