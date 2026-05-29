import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import Client from "../client";

const client = new Client(window.location.origin);

export default function ResultsPanel() {
  return (
    <div className="space-y-8">
      <AssignmentsSection />
      <PacerSection />
    </div>
  );
}

// ---- Assignments ----

function AssignmentsSection() {
  const [runID, setRunID] = useState<number>(0);
  const [inputVal, setInputVal] = useState("");

  const { data, isLoading } = useQuery({
    queryKey: ["assignments", runID],
    queryFn: () => client.domain.ListAssignments({ RunID: runID }),
    enabled: runID > 0,
  });

  return (
    <div className="space-y-4">
      <h2 className="text-lg font-semibold">Assignments</h2>
      <form
        onSubmit={(e) => {
          e.preventDefault();
          setRunID(parseInt(inputVal));
        }}
        className="flex items-end gap-3"
      >
        <div>
          <label className="block text-xs font-medium text-gray-700">Run ID</label>
          <input
            type="number"
            className="mt-1 block w-32 rounded-md border-gray-300 text-sm shadow-sm"
            value={inputVal}
            onChange={(e) => setInputVal(e.target.value)}
            required
          />
        </div>
        <button
          type="submit"
          className="rounded-md bg-black px-3 py-2 text-sm font-medium text-white hover:bg-gray-800"
        >
          Load
        </button>
      </form>

      {runID > 0 && isLoading ? (
        <p className="text-sm text-gray-500">Loading...</p>
      ) : runID > 0 && data ? (
        <>
          <p className="text-sm text-gray-600">
            {data.assignments?.length ?? 0} assignments for run #{runID}
          </p>
          <table className="min-w-full divide-y divide-gray-200 text-sm">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-2 text-left font-medium text-gray-500">ID</th>
                <th className="px-4 py-2 text-left font-medium text-gray-500">User</th>
                <th className="px-4 py-2 text-left font-medium text-gray-500">Lever</th>
                <th className="px-4 py-2 text-left font-medium text-gray-500">Cost</th>
                <th className="px-4 py-2 text-left font-medium text-gray-500">ROI</th>
                <th className="px-4 py-2 text-left font-medium text-gray-500">Status</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {data.assignments?.map((a) => (
                <tr key={a.id}>
                  <td className="px-4 py-2">{a.id}</td>
                  <td className="px-4 py-2">{a.user_id}</td>
                  <td className="px-4 py-2">{a.lever_id}</td>
                  <td className="px-4 py-2">${a.predicted_cost.toFixed(2)}</td>
                  <td className="px-4 py-2">{a.predicted_roi.toFixed(4)}</td>
                  <td className="px-4 py-2">
                    <span className="inline-flex rounded-full bg-green-100 px-2 py-0.5 text-xs font-medium text-green-800">
                      {a.status}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </>
      ) : null}
    </div>
  );
}

// ---- Pacer State ----

function PacerSection() {
  const { data: teamsData } = useQuery({
    queryKey: ["teams"],
    queryFn: () => client.config.ListTeams(),
  });

  const [teamID, setTeamID] = useState<number>(0);

  // Current quarter period
  const now = new Date();
  const quarterMonth = Math.floor(now.getMonth() / 3) * 3;
  const periodStart = new Date(now.getFullYear(), quarterMonth, 1);
  const periodEnd = new Date(now.getFullYear(), quarterMonth + 3, 1);

  const { data, isLoading } = useQuery({
    queryKey: ["pacer", teamID],
    queryFn: () =>
      client.pacer.GetAvailableBudget({
        team_id: teamID,
        period_start: periodStart.toISOString(),
        period_end: periodEnd.toISOString(),
      }),
    enabled: teamID > 0,
  });

  return (
    <div className="space-y-4">
      <h2 className="text-lg font-semibold">Budget Pacing</h2>
      <div>
        <label className="block text-xs font-medium text-gray-700">Team</label>
        <select
          className="mt-1 block w-48 rounded-md border-gray-300 text-sm shadow-sm"
          value={teamID}
          onChange={(e) => setTeamID(Number(e.target.value))}
        >
          <option value={0}>Select team...</option>
          {teamsData?.teams?.map((t) => (
            <option key={t.id} value={t.id}>
              {t.name}
            </option>
          ))}
        </select>
      </div>

      {teamID > 0 && isLoading ? (
        <p className="text-sm text-gray-500">Loading...</p>
      ) : teamID > 0 && data ? (
        <>
          <p className="text-sm text-gray-600">
            Period: {periodStart.toLocaleDateString()} -{" "}
            {periodEnd.toLocaleDateString()}
          </p>
          <table className="min-w-full divide-y divide-gray-200 text-sm">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-2 text-left font-medium text-gray-500">Lever</th>
                <th className="px-4 py-2 text-left font-medium text-gray-500">Configured</th>
                <th className="px-4 py-2 text-left font-medium text-gray-500">Observed</th>
                <th className="px-4 py-2 text-left font-medium text-gray-500">Liability</th>
                <th className="px-4 py-2 text-left font-medium text-gray-500">Available</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {data.caps?.map((cap) => (
                <tr key={cap.lever_id}>
                  <td className="px-4 py-2">{cap.lever_id}</td>
                  <td className="px-4 py-2">${cap.configured.toFixed(2)}</td>
                  <td className="px-4 py-2">${cap.observed.toFixed(2)}</td>
                  <td className="px-4 py-2">
                    ${cap.predicted_liability.toFixed(2)}
                  </td>
                  <td className="px-4 py-2 font-medium">
                    ${cap.available.toFixed(2)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </>
      ) : null}
    </div>
  );
}
