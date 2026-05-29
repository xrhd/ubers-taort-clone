import { ReactNode } from "react";

interface LayoutProps<T extends string> {
  tabs: readonly T[];
  activeTab: T;
  onTabChange: (tab: T) => void;
  children: ReactNode;
}

export default function Layout<T extends string>({
  tabs,
  activeTab,
  onTabChange,
  children,
}: LayoutProps<T>) {
  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-black text-white">
        <div className="mx-auto max-w-7xl px-4 py-4 sm:px-6 lg:px-8">
          <h1 className="text-xl font-bold tracking-tight">
            TAROT
            <span className="ml-2 text-sm font-normal text-gray-400">
              Targeting &amp; Observation Toolkit
            </span>
          </h1>
        </div>
      </header>

      <nav className="border-b border-gray-200 bg-white">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <div className="flex space-x-8">
            {tabs.map((tab) => (
              <button
                key={tab}
                onClick={() => onTabChange(tab)}
                className={`border-b-2 px-1 py-3 text-sm font-medium ${
                  activeTab === tab
                    ? "border-black text-black"
                    : "border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700"
                }`}
              >
                {tab}
              </button>
            ))}
          </div>
        </div>
      </nav>

      <main className="mx-auto max-w-7xl px-4 py-6 sm:px-6 lg:px-8">
        {children}
      </main>
    </div>
  );
}
