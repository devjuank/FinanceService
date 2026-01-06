"use client";

import { useEffect, useState } from "react";
import UploadButton from "@/components/UploadButton";
import { FlowChart, CategoryChart } from "@/components/Charts";

const MOCK_FLOW = [
  { month: 'Oct', income: 4500, expense: 3200 },
  { month: 'Nov', income: 5200, expense: 4100 },
  { month: 'Dec', income: 4800, expense: 3800 },
  { month: 'Jan', income: 6100, expense: 4200 },
];

const MOCK_CATEGORIES = [
  { name: 'Housing', value: 2500 },
  { name: 'Food', value: 800 },
  { name: 'Transport', value: 400 },
  { name: 'Services', value: 1200 },
];

export default function DashboardPage() {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const token = localStorage.getItem("token");
    if (!token) {
      window.location.href = "/auth/login";
    } else {
      setIsAuthenticated(true);
    }
    setLoading(false);
  }, []);

  if (loading) return <div className="p-10">Loading...</div>;
  if (!isAuthenticated) return null;

  return (
    <div className="flex min-h-screen">
      {/* Sidebar */}
      <aside className="w-80 p-10 flex flex-col gap-10">
        <div className="text-2xl font-bold tracking-tight">Finance AI</div>

        <nav className="flex flex-col gap-2">
          <a href="#" className="flex items-center gap-3 px-4 py-3 rounded-xl bg-white/5 text-p-blue font-medium">
            Overview
          </a>
          <a href="#" className="flex items-center gap-3 px-4 py-3 rounded-xl text-gray-400 hover:bg-white/5 transition-all">
            Transactions
          </a>
          <a href="#" className="flex items-center gap-3 px-4 py-3 rounded-xl text-gray-400 hover:bg-white/5 transition-all">
            Analytics
          </a>
        </nav>

        <div className="mt-auto p-6 rounded-2xl glass flex flex-col gap-4">
          <div className="text-sm font-medium text-gray-400">Data Quality</div>
          <div className="text-2xl font-bold">94.2%</div>
          <div className="h-1.5 w-full bg-white/5 rounded-full overflow-hidden">
            <div className="h-full bg-p-blue w-[94.2%]"></div>
          </div>
          <button
            onClick={() => { localStorage.removeItem("token"); window.location.href = "/auth/login"; }}
            className="text-xs text-red-400 font-medium hover:underline text-left"
          >
            Logout
          </button>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 p-10 max-w-6xl">
        <header className="flex justify-between items-end mb-12">
          <div>
            <span className="text-xs font-bold uppercase tracking-widest text-p-blue mb-2 block">Smart Analysis</span>
            <h1 className="text-5xl font-bold tracking-tight">Intelligence Dashboard</h1>
          </div>
          <UploadButton />
        </header>

        <div className="grid grid-cols-3 gap-6 mb-12">
          <div className="card-premium">
            <div className="text-sm font-medium text-gray-400 mb-1">Total Income</div>
            <div className="text-3xl font-bold text-p-green">$ 12,450.00</div>
          </div>
          <div className="card-premium">
            <div className="text-sm font-medium text-gray-400 mb-1">Total Expenses</div>
            <div className="text-3xl font-bold text-p-red">$ 8,120.30</div>
          </div>
          <div className="card-premium">
            <div className="text-sm font-medium text-gray-400 mb-1">Net Savings</div>
            <div className="text-3xl font-bold">$ 4,329.70</div>
            <div className="text-xs font-semibold text-p-green mt-2">+34.8% vs last month</div>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-6">
          <div className="card-premium h-80 flex flex-col gap-4">
            <h2 className="text-lg font-semibold">Income vs Expenses</h2>
            <div className="flex-1">
              <FlowChart data={MOCK_FLOW} />
            </div>
          </div>
          <div className="card-premium h-80 flex flex-col gap-4">
            <h2 className="text-lg font-semibold">Expenses by Category</h2>
            <div className="flex-1">
              <CategoryChart data={MOCK_CATEGORIES} />
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}
