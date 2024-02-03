export const reactQueryKeys = {
  layers: ["layers"],
  repositories: ["repositories"],
  attempts: (runId: string) => ["run", runId, "attempts"],
  logs: (runId: string, attemptId: number | null) => ["logs", runId, attemptId],
};
