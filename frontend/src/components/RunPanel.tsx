import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import Client from "../client";

const client = new Client(window.location.origin);

export default function RunPanel() {
  const qc = useQueryClient();
  const [teamID, setTeamID] = useState<number>(0);
  const [cohortID, setCohortID] = useState<number>(0);
  const [lastRunID, setLastRunID] = useState<number | null>(null);

  const { data: teamsData } = useQuery({
    queryKey: ["teams"],
    queryFn: () => client.config.ListTeams(),
  });

  const { data: cohortsData } = useQuery({
    queryKey: ["cohorts"],
    queryFn: () => client.config.ListCohorts(),
  });

  const syncBudgets = useMutation({
    mutationFn: () => client.budgeting.SyncFromConfig(),
  });

  const triggerRun = useMutation({
    mutationFn: () =>
      client.orchestrator.TriggerRun({ team_id: teamID, cohort_id: cohortID }),
    onSuccess: (resp) => {
      setLastRunID(resp.run_id);
      qc.invalidateQueries({ queryKey: ["runStatus"] });
    },
  });

  const { data: runStatus } = useQuery({
    queryKey: ["runStatus", lastRunID],
    queryFn: () => client.orchestrator.GetRunStatus(lastRunID!),
    enabled: lastRunID !== null,
    refetchInterval: (query) => {
      const status = query.state.data?.status;
      return status === "completed" || status === "failed" ? false : 1000;
    },
  });

  return (
    <div className="space-y-6">
      <h2 className="text-lg font-semibold">Targeting Runs</h2>

      {/* Sync Budgets */}
      <div>
        <button
          onClick={() => syncBudgets.mutate()}
          disabled={syncBudgets.isPending}
          className="rounded-md bg-gray-700 px-3 py-2 text-sm font-medium text-white hover:bg-gray-600 disabled:opacity-50"
        >
          {syncBudgets.isPending ? "Syncing..." : "Sync Budgets from Config"}
        </button>
        {syncBudgets.isSuccess && (
          <span className="ml-2 text-sm text-green-600">Synced!</span>
        )}
      </div>

      {/* Trigger Run */}
      <form
        onSubmit={(e) => {
          e.preventDefault();
          triggerRun.mutate();
        }}
        className="flex items-end gap-3"
      >
        <div>
          <label className="block text-xs font-medium text-gray-700">Team</label>
          <select
            className="mt-1 block w-48 rounded-md border-gray-300 text-sm shadow-sm"
            value={teamID}
            onChange={(e) => setTeamID(Number(e.target.value))}
            required
          >
            <option value={0}>Select team...</option>
            {teamsData?.teams?.map((t) => (
              <option key={t.id} value={t.id}>
                {t.name}
              </option>
            ))}
          </select>
        </div>
        <div>
          <label className="block text-xs font-medium text-gray-700">Cohort</label>
          <select
            className="mt-1 block w-48 rounded-md border-gray-300 text-sm shadow-sm"
            value={cohortID}
            onChange={(e) => setCohortID(Number(e.target.value))}
            required
          >
            <option value={0}>Select cohort...</option>
            {cohortsData?.cohorts?.map((c) => (
              <option key={c.id} value={c.id}>
                {c.name}
              </option>
            ))}
          </select>
        </div>
        <button
          type="submit"
          disabled={triggerRun.isPending || teamID === 0 || cohortID === 0}
          className="rounded-md bg-black px-4 py-2 text-sm font-medium text-white hover:bg-gray-800 disabled:opacity-50"
        >
          {triggerRun.isPending ? "Running..." : "Trigger Run"}
        </button>
      </form>

      {triggerRun.isError && (
        <p className="text-sm text-red-600">
          Error: {(triggerRun.error as Error).message}
        </p>
      )}

      {/* Run Status */}
      {runStatus && (
        <div className="rounded-lg border border-gray-200 bg-white p-4">
          <div className="mb-3 flex items-center gap-3">
            <h3 className="font-medium">Run #{runStatus.id}</h3>
            <StatusBadge status={runStatus.status} />
          </div>

          {runStatus.error && (
            <p className="mb-3 text-sm text-red-600">{runStatus.error}</p>
          )}

          {/* Steps Timeline */}
          {runStatus.steps && runStatus.steps.length > 0 && (
            <div className="space-y-2">
              <h4 className="text-sm font-medium text-gray-700">Pipeline Steps</h4>
              <div className="space-y-1">
                {runStatus.steps.map((step, i) => {
                  const duration =
                    step.finished_at && step.started_at
                      ? new Date(step.finished_at).getTime() -
                        new Date(step.started_at).getTime()
                      : null;

                  return (
                    <div
                      key={i}
                      className="flex items-center gap-3 rounded bg-gray-50 px-3 py-1.5 text-sm"
                    >
                      <StatusDot status={step.status} />
                      <span className="w-28 font-mono">{step.step}</span>
                      <StatusBadge status={step.status} />
                      {duration !== null && (
                        <span className="text-xs text-gray-500">
                          {duration}ms
                        </span>
                      )}
                    </div>
                  );
                })}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  const colors: Record<string, string> = {
    completed: "bg-green-100 text-green-800",
    failed: "bg-red-100 text-red-800",
    started: "bg-yellow-100 text-yellow-800",
    running: "bg-blue-100 text-blue-800",
  };
  return (
    <span
      className={`inline-flex rounded-full px-2 py-0.5 text-xs font-medium ${
        colors[status] || "bg-gray-100 text-gray-800"
      }`}
    >
      {status}
    </span>
  );
}

function StatusDot({ status }: { status: string }) {
  const colors: Record<string, string> = {
    completed: "bg-green-500",
    failed: "bg-red-500",
    started: "bg-yellow-500",
  };
  return (
    <span
      className={`inline-block h-2 w-2 rounded-full ${
        colors[status] || "bg-gray-400"
      }`}
    />
  );
}
