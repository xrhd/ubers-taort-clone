import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import Client, { config } from "../client";

const client = new Client(window.location.origin);

type Section = "teams" | "levers" | "budgets" | "cohorts";

export default function ConfigPanel() {
  const [section, setSection] = useState<Section>("teams");

  return (
    <div>
      <div className="mb-4 flex space-x-2">
        {(["teams", "levers", "budgets", "cohorts"] as Section[]).map((s) => (
          <button
            key={s}
            onClick={() => setSection(s)}
            className={`rounded-md px-3 py-1.5 text-sm font-medium ${
              section === s
                ? "bg-black text-white"
                : "bg-gray-200 text-gray-700 hover:bg-gray-300"
            }`}
          >
            {s.charAt(0).toUpperCase() + s.slice(1)}
          </button>
        ))}
      </div>

      {section === "teams" && <TeamsSection />}
      {section === "levers" && <LeversSection />}
      {section === "budgets" && <BudgetsSection />}
      {section === "cohorts" && <CohortsSection />}
    </div>
  );
}

// ---- Teams ----

function TeamsSection() {
  const qc = useQueryClient();
  const { data, isLoading } = useQuery({
    queryKey: ["teams"],
    queryFn: () => client.config.ListTeams(),
  });

  const [name, setName] = useState("");
  const [lob, setLob] = useState("");

  const create = useMutation({
    mutationFn: (p: config.CreateTeamParams) => client.config.CreateTeam(p),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["teams"] });
      setName("");
      setLob("");
    },
  });

  return (
    <div className="space-y-4">
      <h2 className="text-lg font-semibold">Teams</h2>
      <form
        onSubmit={(e) => {
          e.preventDefault();
          create.mutate({ name, line_of_business: lob });
        }}
        className="flex items-end gap-3"
      >
        <div>
          <label className="block text-xs font-medium text-gray-700">Name</label>
          <input
            className="mt-1 block w-48 rounded-md border-gray-300 text-sm shadow-sm"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
          />
        </div>
        <div>
          <label className="block text-xs font-medium text-gray-700">Line of Business</label>
          <input
            className="mt-1 block w-48 rounded-md border-gray-300 text-sm shadow-sm"
            value={lob}
            onChange={(e) => setLob(e.target.value)}
            required
          />
        </div>
        <button
          type="submit"
          disabled={create.isPending}
          className="rounded-md bg-black px-3 py-2 text-sm font-medium text-white hover:bg-gray-800 disabled:opacity-50"
        >
          Create
        </button>
      </form>

      {isLoading ? (
        <p className="text-sm text-gray-500">Loading...</p>
      ) : (
        <table className="min-w-full divide-y divide-gray-200 text-sm">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-2 text-left font-medium text-gray-500">ID</th>
              <th className="px-4 py-2 text-left font-medium text-gray-500">Name</th>
              <th className="px-4 py-2 text-left font-medium text-gray-500">Line of Business</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {data?.teams?.map((t) => (
              <tr key={t.id}>
                <td className="px-4 py-2">{t.id}</td>
                <td className="px-4 py-2">{t.name}</td>
                <td className="px-4 py-2">{t.line_of_business}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}

// ---- Levers ----

function LeversSection() {
  const qc = useQueryClient();
  const { data: teamsData } = useQuery({
    queryKey: ["teams"],
    queryFn: () => client.config.ListTeams(),
  });

  const [teamID, setTeamID] = useState<number>(0);
  const [name, setName] = useState("");
  const [desc, setDesc] = useState("");

  const { data, isLoading } = useQuery({
    queryKey: ["levers", teamID],
    queryFn: () => client.config.ListLevers({ TeamID: teamID }),
    enabled: teamID > 0,
  });

  const create = useMutation({
    mutationFn: (p: config.CreateLeverParams) => client.config.CreateLever(p),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["levers"] });
      setName("");
      setDesc("");
    },
  });

  return (
    <div className="space-y-4">
      <h2 className="text-lg font-semibold">Levers</h2>

      <div className="flex items-end gap-3">
        <div>
          <label className="block text-xs font-medium text-gray-700">Filter by Team</label>
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
      </div>

      {teamID > 0 && (
        <form
          onSubmit={(e) => {
            e.preventDefault();
            create.mutate({
              team_id: teamID,
              name,
              description: desc,
              eligibility_predicate: {},
            });
          }}
          className="flex items-end gap-3"
        >
          <div>
            <label className="block text-xs font-medium text-gray-700">Name</label>
            <input
              className="mt-1 block w-48 rounded-md border-gray-300 text-sm shadow-sm"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-700">Description</label>
            <input
              className="mt-1 block w-64 rounded-md border-gray-300 text-sm shadow-sm"
              value={desc}
              onChange={(e) => setDesc(e.target.value)}
            />
          </div>
          <button
            type="submit"
            disabled={create.isPending}
            className="rounded-md bg-black px-3 py-2 text-sm font-medium text-white hover:bg-gray-800 disabled:opacity-50"
          >
            Create
          </button>
        </form>
      )}

      {teamID > 0 && isLoading ? (
        <p className="text-sm text-gray-500">Loading...</p>
      ) : (
        teamID > 0 && (
          <table className="min-w-full divide-y divide-gray-200 text-sm">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-2 text-left font-medium text-gray-500">ID</th>
                <th className="px-4 py-2 text-left font-medium text-gray-500">Name</th>
                <th className="px-4 py-2 text-left font-medium text-gray-500">Description</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {data?.levers?.map((l) => (
                <tr key={l.id}>
                  <td className="px-4 py-2">{l.id}</td>
                  <td className="px-4 py-2">{l.name}</td>
                  <td className="px-4 py-2">{l.description}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )
      )}
    </div>
  );
}

// ---- Budget Configs ----

function BudgetsSection() {
  const qc = useQueryClient();
  const { data: teamsData } = useQuery({
    queryKey: ["teams"],
    queryFn: () => client.config.ListTeams(),
  });

  const [teamID, setTeamID] = useState<number>(0);
  const [leverID, setLeverID] = useState<number>(0);
  const [periodStart, setPeriodStart] = useState("");
  const [periodEnd, setPeriodEnd] = useState("");
  const [total, setTotal] = useState("");

  const { data: leversData } = useQuery({
    queryKey: ["levers", teamID],
    queryFn: () => client.config.ListLevers({ TeamID: teamID }),
    enabled: teamID > 0,
  });

  const { data, isLoading } = useQuery({
    queryKey: ["budgetConfigs", teamID],
    queryFn: () => client.config.ListBudgetConfigs({ TeamID: teamID }),
    enabled: teamID > 0,
  });

  const create = useMutation({
    mutationFn: (p: config.CreateBudgetConfigParams) =>
      client.config.CreateBudgetConfig(p),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["budgetConfigs"] });
      setLeverID(0);
      setPeriodStart("");
      setPeriodEnd("");
      setTotal("");
    },
  });

  return (
    <div className="space-y-4">
      <h2 className="text-lg font-semibold">Budget Configs</h2>

      <div>
        <label className="block text-xs font-medium text-gray-700">Filter by Team</label>
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

      {teamID > 0 && (
        <form
          onSubmit={(e) => {
            e.preventDefault();
            create.mutate({
              team_id: teamID,
              lever_id: leverID,
              period_start: new Date(periodStart).toISOString(),
              period_end: new Date(periodEnd).toISOString(),
              configured_total: parseFloat(total),
            });
          }}
          className="flex flex-wrap items-end gap-3"
        >
          <div>
            <label className="block text-xs font-medium text-gray-700">Lever</label>
            <select
              className="mt-1 block w-48 rounded-md border-gray-300 text-sm shadow-sm"
              value={leverID}
              onChange={(e) => setLeverID(Number(e.target.value))}
              required
            >
              <option value={0}>Select lever...</option>
              {leversData?.levers?.map((l) => (
                <option key={l.id} value={l.id}>
                  {l.name}
                </option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-700">Period Start</label>
            <input
              type="date"
              className="mt-1 block w-40 rounded-md border-gray-300 text-sm shadow-sm"
              value={periodStart}
              onChange={(e) => setPeriodStart(e.target.value)}
              required
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-700">Period End</label>
            <input
              type="date"
              className="mt-1 block w-40 rounded-md border-gray-300 text-sm shadow-sm"
              value={periodEnd}
              onChange={(e) => setPeriodEnd(e.target.value)}
              required
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-700">Total ($)</label>
            <input
              type="number"
              step="0.01"
              className="mt-1 block w-32 rounded-md border-gray-300 text-sm shadow-sm"
              value={total}
              onChange={(e) => setTotal(e.target.value)}
              required
            />
          </div>
          <button
            type="submit"
            disabled={create.isPending}
            className="rounded-md bg-black px-3 py-2 text-sm font-medium text-white hover:bg-gray-800 disabled:opacity-50"
          >
            Create
          </button>
        </form>
      )}

      {teamID > 0 && isLoading ? (
        <p className="text-sm text-gray-500">Loading...</p>
      ) : (
        teamID > 0 && (
          <table className="min-w-full divide-y divide-gray-200 text-sm">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-2 text-left font-medium text-gray-500">ID</th>
                <th className="px-4 py-2 text-left font-medium text-gray-500">Lever</th>
                <th className="px-4 py-2 text-left font-medium text-gray-500">Period</th>
                <th className="px-4 py-2 text-left font-medium text-gray-500">Total</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {data?.budget_configs?.map((b) => (
                <tr key={b.id}>
                  <td className="px-4 py-2">{b.id}</td>
                  <td className="px-4 py-2">{b.lever_id}</td>
                  <td className="px-4 py-2">
                    {new Date(b.period_start).toLocaleDateString()} -{" "}
                    {new Date(b.period_end).toLocaleDateString()}
                  </td>
                  <td className="px-4 py-2">${b.configured_total.toFixed(2)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )
      )}
    </div>
  );
}

// ---- Cohorts ----

function CohortsSection() {
  const qc = useQueryClient();
  const { data, isLoading } = useQuery({
    queryKey: ["cohorts"],
    queryFn: () => client.config.ListCohorts(),
  });

  const [name, setName] = useState("");
  const [predicate, setPredicate] = useState("{}");

  const create = useMutation({
    mutationFn: (p: config.CreateCohortParams) => client.config.CreateCohort(p),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["cohorts"] });
      setName("");
      setPredicate("{}");
    },
  });

  return (
    <div className="space-y-4">
      <h2 className="text-lg font-semibold">Cohorts</h2>
      <form
        onSubmit={(e) => {
          e.preventDefault();
          try {
            create.mutate({ name, predicate: JSON.parse(predicate) });
          } catch {
            alert("Invalid JSON predicate");
          }
        }}
        className="flex items-end gap-3"
      >
        <div>
          <label className="block text-xs font-medium text-gray-700">Name</label>
          <input
            className="mt-1 block w-48 rounded-md border-gray-300 text-sm shadow-sm"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
          />
        </div>
        <div>
          <label className="block text-xs font-medium text-gray-700">
            Predicate (JSON)
          </label>
          <input
            className="mt-1 block w-64 rounded-md border-gray-300 text-sm shadow-sm font-mono"
            value={predicate}
            onChange={(e) => setPredicate(e.target.value)}
            required
          />
        </div>
        <button
          type="submit"
          disabled={create.isPending}
          className="rounded-md bg-black px-3 py-2 text-sm font-medium text-white hover:bg-gray-800 disabled:opacity-50"
        >
          Create
        </button>
      </form>

      {isLoading ? (
        <p className="text-sm text-gray-500">Loading...</p>
      ) : (
        <table className="min-w-full divide-y divide-gray-200 text-sm">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-2 text-left font-medium text-gray-500">ID</th>
              <th className="px-4 py-2 text-left font-medium text-gray-500">Name</th>
              <th className="px-4 py-2 text-left font-medium text-gray-500">Predicate</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {data?.cohorts?.map((c) => (
              <tr key={c.id}>
                <td className="px-4 py-2">{c.id}</td>
                <td className="px-4 py-2">{c.name}</td>
                <td className="px-4 py-2 font-mono text-xs">
                  {JSON.stringify(c.predicate)}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
