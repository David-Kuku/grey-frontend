import { Link, Outlet, useLocation } from "react-router-dom";
import { useLogout } from "../queries/auth.queries";

const NAV_LINKS = [
  { to: "/", label: "Dashboard" },
  { to: "/deposit", label: "Deposit" },
  { to: "/convert", label: "Convert" },
  { to: "/payout", label: "Payout" },
  { to: "/transactions", label: "Transactions" },
];

const Layout = () => {
  const logout = useLogout();
  const { pathname } = useLocation();

  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="bg-white border-b border-gray-200">
        <div className="max-w-5xl mx-auto px-4 flex items-center justify-between h-14">
          <div className="flex items-center gap-6">
            <span className="font-semibold text-gray-900 text-lg">Kite</span>
            <div className="flex gap-1">
              {NAV_LINKS.map(({ to, label }) => (
                <Link
                  key={to}
                  to={to}
                  className={`px-3 py-1.5 rounded text-sm font-medium transition-colors ${
                    pathname === to
                      ? "bg-gray-100 text-gray-900"
                      : "text-gray-500 hover:text-gray-900"
                  }`}
                >
                  {label}
                </Link>
              ))}
            </div>
          </div>
          <button
            onClick={logout}
            className="text-sm text-gray-500 hover:text-gray-900"
          >
            Sign out
          </button>
        </div>
      </nav>
      <main className="max-w-5xl mx-auto px-4 py-8">
        <Outlet />
      </main>
    </div>
  );
};

export default Layout;
