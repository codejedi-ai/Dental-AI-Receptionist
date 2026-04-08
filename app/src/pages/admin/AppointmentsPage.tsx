import { useEffect, useState } from "react";
import { formatTime } from "@/lib/utils";

interface Appointment { id: number; patient_name: string; patient_email: string; dentist: string; service: string; date: string; time: string; status: string; }

export default function AppointmentsPage() {
  const [appointments, setAppointments] = useState<Appointment[]>([]);
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");
  const [loading, setLoading] = useState(true);

  const fetchAppts = () => {
    setLoading(true);
    const p = new URLSearchParams();
    if (search) p.set("search", search);
    if (statusFilter !== "all") p.set("status", statusFilter);
    fetch(`/api/appointments?${p}`).then((r) => r.json()).then((d) => { setAppointments(d.appointments || []); setLoading(false); }).catch(() => setLoading(false));
  };

  useEffect(() => { fetchAppts(); }, [statusFilter]);

  const updateStatus = async (id: number, s: string) => {
    await fetch(`/api/appointments/${id}`, { method: "PATCH", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ status: s }) });
    fetchAppts();
  };

  const badge = (s: string) => {
    const c: Record<string, string> = { confirmed: "bg-green-100 text-green-800", completed: "bg-blue-100 text-blue-800", cancelled: "bg-red-100 text-red-800" };
    return <span className={`px-2.5 py-1 rounded-full text-xs font-semibold ${c[s] || "bg-gray-100 text-gray-800"}`}>{s}</span>;
  };

  return (
    <div>
      <h1 className="text-2xl font-bold text-gray-900 mb-6">Appointments</h1>
      <div className="flex flex-col sm:flex-row gap-3 mb-6">
        <form onSubmit={(e) => { e.preventDefault(); fetchAppts(); }} className="flex-1 flex gap-2">
          <input type="text" value={search} onChange={(e) => setSearch(e.target.value)} placeholder="Search patient, service, or dentist..." className="flex-1 rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-teal-500" />
          <button type="submit" className="px-4 py-2 bg-teal-600 text-white rounded-lg text-sm font-medium hover:bg-teal-700">Search</button>
        </form>
        <select value={statusFilter} onChange={(e) => setStatusFilter(e.target.value)} className="rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-teal-500">
          <option value="all">All Status</option><option value="confirmed">Confirmed</option><option value="completed">Completed</option><option value="cancelled">Cancelled</option>
        </select>
      </div>
      <div className="bg-white rounded-xl border border-gray-200 overflow-hidden">
        {loading ? <div className="p-8 text-center text-gray-400">Loading...</div> : appointments.length === 0 ? <div className="p-8 text-center text-gray-400"><span className="text-4xl block mb-2">📭</span><p>No appointments found</p></div> : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead><tr className="bg-gray-50 border-b border-gray-200">
                {["#", "Patient", "Service", "Dentist", "Date", "Time", "Status", "Actions"].map((h) => <th key={h} className="text-left px-4 py-3 font-semibold text-gray-600">{h}</th>)}
              </tr></thead>
              <tbody className="divide-y divide-gray-100">
                {appointments.map((a) => (
                  <tr key={a.id} className="hover:bg-gray-50">
                    <td className="px-4 py-3 text-gray-500">#{a.id}</td>
                    <td className="px-4 py-3"><p className="font-medium text-gray-900">{a.patient_name || "—"}</p><p className="text-xs text-gray-500">{a.patient_email}</p></td>
                    <td className="px-4 py-3 text-gray-700">{a.service}</td>
                    <td className="px-4 py-3 text-gray-700">{a.dentist}</td>
                    <td className="px-4 py-3 text-gray-700">{a.date}</td>
                    <td className="px-4 py-3 text-gray-700">{formatTime(a.time)}</td>
                    <td className="px-4 py-3">{badge(a.status)}</td>
                    <td className="px-4 py-3">
                      <div className="flex gap-1">
                        {a.status === "confirmed" && (<><button onClick={() => updateStatus(a.id, "completed")} className="px-2 py-1 bg-blue-50 text-blue-700 rounded text-xs font-medium hover:bg-blue-100">Complete</button><button onClick={() => updateStatus(a.id, "cancelled")} className="px-2 py-1 bg-red-50 text-red-700 rounded text-xs font-medium hover:bg-red-100">Cancel</button></>)}
                        {a.status === "cancelled" && <button onClick={() => updateStatus(a.id, "confirmed")} className="px-2 py-1 bg-green-50 text-green-700 rounded text-xs font-medium hover:bg-green-100">Reconfirm</button>}
                        {a.status === "completed" && <span className="text-xs text-gray-400">Done</span>}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
      <p className="mt-4 text-sm text-gray-500">Showing {appointments.length} appointment{appointments.length !== 1 ? "s" : ""}</p>
    </div>
  );
}
