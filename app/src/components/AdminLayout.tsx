import { Link, Outlet, useLocation } from "react-router-dom";

const navItems = [
  { to: "/admin", label: "Dashboard", icon: "📊" },
  { to: "/admin/appointments", label: "Appointments", icon: "📅" },
];

export default function AdminLayout() {
  const { pathname } = useLocation();

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200 sticky top-0 z-40">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 flex items-center justify-between h-14">
          <div className="flex items-center space-x-4">
            <Link to="/admin" className="flex items-center space-x-2">
              <span className="text-2xl">🦷</span>
              <span className="font-bold text-gray-900">Smile Dental</span>
              <span className="text-xs bg-teal-100 text-teal-700 px-2 py-0.5 rounded-full font-medium">Admin</span>
            </Link>
          </div>
          <Link to="/" className="text-sm text-gray-500 hover:text-gray-700">← Back to site</Link>
        </div>
      </header>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <nav className="flex space-x-1 mb-6 bg-white rounded-xl p-1 border border-gray-200 w-fit">
          {navItems.map((item) => (
            <Link key={item.to} to={item.to} className={`flex items-center space-x-2 px-4 py-2 rounded-lg text-sm font-medium transition-colors ${pathname === item.to ? "bg-teal-600 text-white" : "text-gray-600 hover:bg-gray-100"}`}>
              <span>{item.icon}</span><span>{item.label}</span>
            </Link>
          ))}
        </nav>
        <Outlet />
      </div>
    </div>
  );
}
