import { useState } from "react";
import Layout from "./components/Layout";
import ConfigPanel from "./components/ConfigPanel";
import RunPanel from "./components/RunPanel";
import ResultsPanel from "./components/ResultsPanel";

const tabs = ["Configuration", "Targeting Runs", "Results"] as const;
type Tab = (typeof tabs)[number];

export default function App() {
  const [activeTab, setActiveTab] = useState<Tab>("Configuration");

  return (
    <Layout tabs={tabs} activeTab={activeTab} onTabChange={setActiveTab}>
      {activeTab === "Configuration" && <ConfigPanel />}
      {activeTab === "Targeting Runs" && <RunPanel />}
      {activeTab === "Results" && <ResultsPanel />}
    </Layout>
  );
}
