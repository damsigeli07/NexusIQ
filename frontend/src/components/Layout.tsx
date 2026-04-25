import { Outlet, NavLink, useNavigate } from "react-router-dom";
import { clearAuth } from "../store/auth";
import { MessageSquare, FileText, BarChart2, LogOut } from "lucide-react";

export default function Layout() {
  const navigate = useNavigate();

  const logout = () => {
    clearAuth();
    navigate("/login");
  };

  const link = "flex items-center gap-2 px-3 py-2 rounded-lg text-sm text-gray-300 hover:bg-gray-700 hover:text-white transition-colors";
  const active = "bg-gray-700 text-white";

  return (
    <div className="flex h-screen bg-gray-950 text-white">
      {/* Sidebar */}
      <aside className="w-56 bg-gray-900 flex flex-col border-r border-gray-800">
        <div className="px-4 py-5 border-b border-gray-800">
          <span className="text-lg font-semibold tracking-tight">
            Nexus<span className="text-indigo-400">IQ</span>
          </span>
        </div>
        <nav className="flex-1 p-3 space-y-1">
          <NavLink to="/chat"
            className={({ isActive }) => `${link} ${isActive ? active : ""}`}>
            <MessageSquare size={16} /> Chat
          </NavLink>
          <NavLink to="/documents"
            className={({ isActive }) => `${link} ${isActive ? active : ""}`}>
            <FileText size={16} /> Documents
          </NavLink>
          <NavLink to="/analytics"
            className={({ isActive }) => `${link} ${isActive ? active : ""}`}>
            <BarChart2 size={16} /> Analytics
          </NavLink>
        </nav>
        <div className="p-3 border-t border-gray-800">
          <button onClick={logout}
            className="flex items-center gap-2 w-full px-3 py-2 text-sm text-gray-400 hover:text-white rounded-lg hover:bg-gray-700 transition-colors">
            <LogOut size={16} /> Sign out
          </button>
        </div>
      </aside>

      {/* Main */}
      <main className="flex-1 overflow-hidden">
        <Outlet />
      </main>
    </div>
  );
}