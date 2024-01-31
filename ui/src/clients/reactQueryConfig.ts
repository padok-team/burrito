export const reactQueryKeys = {
  layers: ["layers"],
  repositories: ["repositories"],
  attempts: (runId: string) => ["run", runId, "attempts"],
};
