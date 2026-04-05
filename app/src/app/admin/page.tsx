"use client";

import { useEffect, useState } from "react";
import { formatTime } from "@/lib/utils";

interface Appointment {
  id: number;
  patient_name: string;
  dentist: string;
  service: string;
  date: string;
  time: string;
  status: string;
}

interface Stats {
  totalPatients: number;
  totalAppointments: number;
  todayAppointments: number;
  weekAppointments: number;
  confirmedCount: number;
  cancelledCount: number;
}

export default function AdminDashboard() {
  const [todayAppts, setTodayAppts] = useState<Appointment[]>([]);
  const [stats, setStats] = useState<Stats>({
    totalPatients: 0,
    totalAppointments: 0,
    todayAppointments: 0,
    weekAppointments: 0,
    confirmedCount: 0,
    cancelledCount: 0,
  });

  const today = new Date().toISOString().split("T")[0];

  useEffect(() => {
    // Fetch today's appointments
    fetch(`/api/appointments?date=${today}`)
      .then((r) => r.json())
      .then((data) => {
        setTodayAppts(data.appointments || []);
        setStats((s) => ({ ...s, todayAppointments: (data.appointments || []).length }));
      });

    // Fetch all appointments for stats
    fetch("/api/appointments")
      .then((r) => r.json())
      .then((data) => {
        const appts = data.appointments || [];
        const now = new Date();
        const weekFromNow = new Date(now);
        weekFromNow.setDate(weekFromNow.getDate() + 7);
        const weekStr = weekFromNow.toISOString().split("T")[0];

        setStats((s) => ({
          ...s,
          totalAppointments: appts.length,
          weekAppointments: appts.filter((a: any) => a.date >= today && a.date <= weekStr).length,
          confirmedCount: appts.filter((a: any) => a.status === "confirmed").length,
          cancelledCount: appts.filter((a: any) => a.status === "cancelled").length,
        }));
      });

    // Fetch patients count
    fetch("/api/patients")
      .then((r) => r.json())
      .then((data) => {
        setStats((s) => ({ ...s, totalPatients: (data.patients || []).length }));
      });
  }, [today]);

  const statCards = [
    { label: "Total Patients", value: stats.totalPatients, icon: "👥", color: "bg-blue-50 text-blue-700" },
    { label: "Today's Appointments", value: stats.todayAppointments, icon: "📅", color: "bg-teal-50 text-teal-700" },
    { label: "This Week", value: stats.weekAppointments, icon: "📊", color: "bg-purple-50 text-purple-700" },
    { label: "Confirmed", value: stats.confirmedCount, icon: "✅", color: "bg-green-50 text-green-700" },
    { label: "Cancelled", value: stats.cancelledCount, icon: "❌", color: "bg-red-50 text-red-700" },
    { label: "Total Appointments", value: stats.totalAppointments, icon: "🗓️", color: "bg-orange-50 text-orange-700" },
  ];

  return (
    <div>
      <h1 className="text-2xl font-bold text-gray-900 mb-6">Dashboard</h1>

      {/* Stats Grid */}
      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-4 mb-8">
        {statCards.map((stat) => (
          <div key={stat.label} className="bg-white rounded-xl border border-gray-200 p-4">
            <div className={`w-10 h-10 rounded-lg ${stat.color} flex items-center justify-center mb-2`}>
              <span className="text-lg">{stat.icon}</span>
            </div>
            <p className="text-2xl font-bold text-gray-900">{stat.value}</p>
            <p className="text-xs text-gray-500">{stat.label}</p>
          </div>
        ))}
      </div>

      {/* Today's Appointments */}
      <div className="bg-white rounded-xl border border-gray-200 overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-100">
          <h2 className="text-lg font-bold text-gray-900">Today&apos;s Appointments</h2>
          <p className="text-sm text-gray-500">{today}</p>
        </div>
        {todayAppts.length === 0 ? (
          <div className="p-8 text-center text-gray-400">
            <span className="text-4xl block mb-2">📭</span>
            <p>No appointments scheduled for today</p>
          </div>
        ) : (
          <div className="divide-y divide-gray-100">
            {todayAppts.map((appt) => (
              <div key={appt.id} className="px-6 py-4 flex items-center justify-between hover:bg-gray-50">
                <div className="flex items-center space-x-4">
                  <div className="text-center bg-teal-50 rounded-lg px-3 py-1.5">
                    <p className="text-sm font-bold text-teal-700">{formatTime(appt.time)}</p>
                  </div>
                  <div>
                    <p className="font-medium text-gray-900">{appt.patient_name || "Patient"}</p>
                    <p className="text-sm text-gray-500">
                      {appt.service} · {appt.dentist}
                    </p>
                  </div>
                </div>
                <span
                  className={`px-2.5 py-1 rounded-full text-xs font-semibold ${
                    appt.status === "confirmed"
                      ? "bg-green-100 text-green-800"
                      : appt.status === "completed"
                      ? "bg-blue-100 text-blue-800"
                      : "bg-red-100 text-red-800"
                  }`}
                >
                  {appt.status}
                </span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
