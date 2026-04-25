import { useEffect, useState } from "react";
import api from "../api/client";
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, Cell } from "recharts";
import { FileText, MessageSquare, Database } from "lucide-react";

type AnalyticsData = {
  documents:       Record<string, number>;
  total_questions: number;
  total_chunks:    number;
  top_questions:   { question: string; frequency: number }[];
};

export default function Analytics() {
  const [data, setData] = useState<AnalyticsData | null>(null);

  useEffect(() => {
    api.get("/analytics").then(r => setData(r.data));
  }, []);

  if (!data) return (
    <div className="h-full flex items-center justify-center text-gray-500 text-sm">
      Loading analytics…
    </div>
  );

  const totalDocs = Object.values(data.documents).reduce((a, b) => a + b, 0);
  const docStatusData = Object.entries(data.documents).map(([k, v]) => ({ name: k, count: v }));

  return (
    <div className="h-full overflow-y-auto p-6 space-y-6">
      <h2 className="text-lg font-semibold">Analytics</h2>

      {/* Stat cards */}
      <div className="grid grid-cols-3 gap-4">
        {[
          { label: "Total documents", value: totalDocs,            icon: <FileText size={18} /> },
          { label: "Questions asked",  value: data.total_questions, icon: <MessageSquare size={18} /> },
          { label: "Knowledge chunks", value: data.total_chunks,    icon: <Database size={18} /> },
        ].map(s => (
          <div key={s.label} className="bg-gray-900 border border-gray-800 rounded-xl p-4 flex items-center gap-4">
            <div className="text-indigo-400">{s.icon}</div>
            <div>
              <p className="text-2xl font-semibold">{s.value}</p>
              <p className="text-xs text-gray-400">{s.label}</p>
            </div>
          </div>
        ))}
      </div>

      {/* Document status bar chart */}
      <div className="bg-gray-900 border border-gray-800 rounded-xl p-4">
        <p className="text-sm font-medium mb-4">Documents by status</p>
        <ResponsiveContainer width="100%" height={160}>
          <BarChart data={docStatusData} barSize={36}>
            <XAxis dataKey="name" tick={{ fontSize: 12, fill: "#9ca3af" }} />
            <YAxis allowDecimals={false} tick={{ fontSize: 12, fill: "#9ca3af" }} />
            <Tooltip
              contentStyle={{ background: "#111827", border: "1px solid #374151", borderRadius: 8 }}
              labelStyle={{ color: "#f3f4f6" }}
            />
            <Bar dataKey="count" radius={[4, 4, 0, 0]}>
              {docStatusData.map((entry) => (
                <Cell key={entry.name}
                  fill={entry.name === "ready" ? "#6366f1" : entry.name === "failed" ? "#ef4444" : "#374151"} />
              ))}
            </Bar>
          </BarChart>
        </ResponsiveContainer>
      </div>

      {/* Top questions */}
      {data.top_questions.length > 0 && (
        <div className="bg-gray-900 border border-gray-800 rounded-xl p-4">
          <p className="text-sm font-medium mb-3">Top questions</p>
          <div className="space-y-2">
            {data.top_questions.map((q, i) => (
              <div key={i} className="flex items-center justify-between text-sm">
                <span className="text-gray-300 truncate mr-4">{q.question}</span>
                <span className="text-indigo-400 shrink-0">{q.frequency}x</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}