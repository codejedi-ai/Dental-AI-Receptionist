"use client";

import { useEffect, useState } from "react";
import { formatTime, formatDate } from "@/lib/utils";

interface Appointment {
  id: number;
  patient_name: string;
  patient_email: string;
  patient_phone: string;
  dentist: string;
  service: string;
  date: string;
  time: string;
  status: string;
  notes: string;
  created_at: string;
}

export default function AppointmentsPage() {
  const [appointments, setAppointments] = useState<Appointment[]>([]);
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");
  const [loading, setLoading] = useState(true);

  const fetchAppointments = () => {
    setLoading(true);
    const params = new URLSearchParams();
    if (search) params.set("search", search);
    if (statusFilter !== "all") params.set("status", statusFilter);

    fetch(`/api/appointments?${params}`)
      .then((r) => r.json())
      .then((data) => {
        setAppointments(data.appointments || []);
        setLoading(false);
      })
      .catch(() => setLoading(false));
  };

  useEffect(() => {
    fetchAppointments();
  }, [statusFilter]);

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    fetchAppointments();
  };

  const updateStatus = async (id: number, newStatus: string) => {
    await fetch(`/api/appointments/${id}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ status: newStatus }),
    });
    fetchAppointments();
  };

  const statusBadge = (status: string) => {
    const styles: Record<string, string> = {
      confirmed: "bg-green-100 text-green-800",
      completed: "bg-blue-100 text-blue-800",
      cancelled: "bg-red-100 text-red-800",
    };
    return (
      <span className={`px-2.5 py-1 rounded-full text-xs font-semibold ${styles[status] || "bg-gray-100 text-gray-800"}`}>
        {status}
      </span>
    );
  };

  return (
    <div>
      <h1 className="text-2xl font-bold text-gray-900 mb-6">Appointments</h1>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-3 mb-6">
        <form onSubmit={handleSearch} className="flex-1 flex gap-2">
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search by patient, service, or dentist..."
            className="flex-1 rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-teal-500"
          />
          <button
            type="submit"
            className="px-4 py-2 bg-teal-600 text-white rounded-lg text-sm font-medium hover:bg-teal-700"
          >
            Search
          </button>
        </form>
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          className="rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-teal-500"
        >
          <option value="all">All Status</option>
          <option value="confirmed">Confirmed</option>
          <option value="completed">Completed</option>
          <option value="cancelled">Cancelled</option>
        </select>
      </div>

      {/* Table */}
      <div className="bg-white rounded-xl border border-gray-200 overflow-hidden">
        {loading ? (
          <div className="p-8 text-center text-gray-400">Loading...</div>
        ) : appointments.length === 0 ? (
          <div className="p-8 text-center text-gray-400">
            <span className="text-4xl block mb-2">📭</span>
            <p>No appointments found</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="bg-gray-50 border-b border-gray-200">
                  <th className="text-left px-4 py-3 font-semibold text-gray-600">#</th>
                  <th className="text-left px-4 py-3 font-semibold text-gray-600">Patient</th>
                  <th className="text-left px-4 py-3 font-semibold text-gray-600">Service</th>
                  <th className="text-left px-4 py-3 font-semibold text-gray-600">Dentist</th>
                  <th className="text-left px-4 py-3 font-semibold text-gray-600">Date</th>
                  <th className="text-left px-4 py-3 font-semibold text-gray-600">Time</th>
                  <th className="text-left px-4 py-3 font-semibold text-gray-600">Status</th>
                  <th className="text-left px-4 py-3 font-semibold text-gray-600">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {appointments.map((appt) => (
                  <tr key={appt.id} className="hover:bg-gray-50">
                    <td className="px-4 py-3 text-gray-500">#{appt.id}</td>
                    <td className="px-4 py-3">
                      <p className="font-medium text-gray-900">{appt.patient_name || "—"}</p>
                      <p className="text-xs text-gray-500">{appt.patient_email}</p>
                    </td>
                    <td className="px-4 py-3 text-gray-700">{appt.service}</td>
                    <td className="px-4 py-3 text-gray-700">{appt.dentist}</td>
                    <td className="px-4 py-3 text-gray-700">{appt.date}</td>
                    <td className="px-4 py-3 text-gray-700">{formatTime(appt.time)}</td>
                    <td className="px-4 py-3">{statusBadge(appt.status)}</td>
                    <td className="px-4 py-3">
                      <div className="flex gap-1">
                        {appt.status === "confirmed" && (
                          <>
                            <button
                              onClick={() => updateStatus(appt.id, "completed")}
                              className="px-2 py-1 bg-blue-50 text-blue-700 rounded text-xs font-medium hover:bg-blue-100"
                            >
                              Complete
                            </button>
                            <button
                              onClick={() => updateStatus(appt.id, "cancelled")}
                              className="px-2 py-1 bg-red-50 text-red-700 rounded text-xs font-medium hover:bg-red-100"
                            >
                              Cancel
                            </button>
                          </>
                        )}
                        {appt.status === "cancelled" && (
                          <button
                            onClick={() => updateStatus(appt.id, "confirmed")}
                            className="px-2 py-1 bg-green-50 text-green-700 rounded text-xs font-medium hover:bg-green-100"
                          >
                            Reconfirm
                          </button>
                        )}
                        {appt.status === "completed" && (
                          <span className="text-xs text-gray-400">Done</span>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      <p className="mt-4 text-sm text-gray-500">
        Showing {appointments.length} appointment{appointments.length !== 1 ? "s" : ""}
      </p>
    </div>
  );
}
