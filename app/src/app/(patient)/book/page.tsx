"use client";

import { useState, useEffect } from "react";
import { clinicConfig } from "@/lib/clinic-config";
import { formatTime } from "@/lib/utils";

type Step = "service" | "dentist" | "datetime" | "info" | "confirm" | "success";

export default function BookPage() {
  const [step, setStep] = useState<Step>("service");
  const [service, setService] = useState("");
  const [dentist, setDentist] = useState("");
  const [date, setDate] = useState("");
  const [time, setTime] = useState("");
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [phone, setPhone] = useState("");
  const [notes, setNotes] = useState("");
  const [availableSlots, setAvailableSlots] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);
  const [appointmentId, setAppointmentId] = useState<number | null>(null);

  // Get tomorrow as minimum date
  const tomorrow = new Date();
  tomorrow.setDate(tomorrow.getDate() + 1);
  const minDate = tomorrow.toISOString().split("T")[0];

  // Max date: 3 months out
  const maxDate = new Date();
  maxDate.setMonth(maxDate.getMonth() + 3);
  const maxDateStr = maxDate.toISOString().split("T")[0];

  useEffect(() => {
    if (date && dentist) {
      fetch(`/api/appointments/available?date=${date}&dentist=${encodeURIComponent(dentist)}`)
        .then((r) => r.json())
        .then((data) => setAvailableSlots(data.slots || []))
        .catch(() => setAvailableSlots([]));
    }
  }, [date, dentist]);

  const handleSubmit = async () => {
    setLoading(true);
    try {
      const res = await fetch("/api/appointments", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ service, dentist, date, time, name, email, phone, notes }),
      });
      const data = await res.json();
      if (data.appointment) {
        setAppointmentId(data.appointment.id);
        setStep("success");
      }
    } catch (err) {
      alert("Something went wrong. Please try again.");
    } finally {
      setLoading(false);
    }
  };

  const steps: { key: Step; label: string }[] = [
    { key: "service", label: "Service" },
    { key: "dentist", label: "Dentist" },
    { key: "datetime", label: "Date & Time" },
    { key: "info", label: "Your Info" },
    { key: "confirm", label: "Confirm" },
  ];

  const stepIndex = steps.findIndex((s) => s.key === step);

  return (
    <>
      <section className="bg-gradient-to-br from-teal-600 to-teal-800 text-white py-12">
        <div className="max-w-4xl mx-auto px-4">
          <h1 className="text-3xl sm:text-4xl font-bold mb-2">Book an Appointment</h1>
          <p className="text-teal-100">Schedule your visit in just a few steps</p>
        </div>
      </section>

      <section className="py-12">
        <div className="max-w-4xl mx-auto px-4">
          {/* Progress bar */}
          {step !== "success" && (
            <div className="mb-10">
              <div className="flex items-center justify-between mb-2">
                {steps.map((s, i) => (
                  <div key={s.key} className="flex items-center flex-1">
                    <div
                      className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-bold ${
                        i <= stepIndex
                          ? "bg-teal-600 text-white"
                          : "bg-gray-200 text-gray-500"
                      }`}
                    >
                      {i + 1}
                    </div>
                    {i < steps.length - 1 && (
                      <div
                        className={`flex-1 h-1 mx-2 rounded ${
                          i < stepIndex ? "bg-teal-600" : "bg-gray-200"
                        }`}
                      />
                    )}
                  </div>
                ))}
              </div>
              <div className="flex justify-between text-xs text-gray-500">
                {steps.map((s) => (
                  <span key={s.key}>{s.label}</span>
                ))}
              </div>
            </div>
          )}

          {/* Step 1: Service */}
          {step === "service" && (
            <div>
              <h2 className="text-2xl font-bold text-gray-900 mb-6">Select a Service</h2>
              <div className="grid grid-cols-2 sm:grid-cols-3 gap-4">
                {clinicConfig.services.map((s) => (
                  <button
                    key={s.name}
                    onClick={() => {
                      setService(s.name);
                      setStep("dentist");
                    }}
                    className={`p-4 rounded-xl border-2 text-center transition-all hover:shadow-md ${
                      service === s.name
                        ? "border-teal-600 bg-teal-50"
                        : "border-gray-100 hover:border-teal-300"
                    }`}
                  >
                    <span className="text-3xl block mb-2">{s.icon}</span>
                    <span className="font-semibold text-gray-900 text-sm">{s.name}</span>
                    <span className="block text-xs text-gray-500 mt-1">{s.price}</span>
                  </button>
                ))}
              </div>
            </div>
          )}

          {/* Step 2: Dentist */}
          {step === "dentist" && (
            <div>
              <h2 className="text-2xl font-bold text-gray-900 mb-6">Choose Your Dentist</h2>
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                {clinicConfig.dentists.map((d) => (
                  <button
                    key={d.id}
                    onClick={() => {
                      setDentist(d.name);
                      setStep("datetime");
                    }}
                    className={`p-6 rounded-xl border-2 text-center transition-all hover:shadow-md ${
                      dentist === d.name
                        ? "border-teal-600 bg-teal-50"
                        : "border-gray-100 hover:border-teal-300"
                    }`}
                  >
                    <div className="w-16 h-16 bg-teal-100 rounded-full flex items-center justify-center mx-auto mb-3">
                      <span className="text-2xl">👨‍⚕️</span>
                    </div>
                    <h3 className="font-bold text-gray-900">{d.name}</h3>
                    <p className="text-teal-600 text-xs mt-1">{d.specialty}</p>
                  </button>
                ))}
              </div>
              <button
                onClick={() => setStep("service")}
                className="mt-6 text-gray-500 hover:text-gray-700 text-sm"
              >
                ← Back to services
              </button>
            </div>
          )}

          {/* Step 3: Date & Time */}
          {step === "datetime" && (
            <div>
              <h2 className="text-2xl font-bold text-gray-900 mb-6">Pick a Date & Time</h2>
              <div className="mb-6">
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Select Date
                </label>
                <input
                  type="date"
                  min={minDate}
                  max={maxDateStr}
                  value={date}
                  onChange={(e) => {
                    setDate(e.target.value);
                    setTime("");
                  }}
                  className="w-full sm:w-64 rounded-lg border border-gray-300 px-3 py-2.5 focus:outline-none focus:ring-2 focus:ring-teal-500"
                />
              </div>

              {date && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Available Times
                  </label>
                  {availableSlots.length === 0 ? (
                    <p className="text-gray-500">No available slots for this date. Please select another day.</p>
                  ) : (
                    <div className="grid grid-cols-3 sm:grid-cols-5 gap-2">
                      {availableSlots.map((slot) => (
                        <button
                          key={slot}
                          onClick={() => {
                            setTime(slot);
                            setStep("info");
                          }}
                          className={`py-2.5 px-3 rounded-lg text-sm font-medium transition-all ${
                            time === slot
                              ? "bg-teal-600 text-white"
                              : "bg-gray-100 text-gray-700 hover:bg-teal-100 hover:text-teal-700"
                          }`}
                        >
                          {formatTime(slot)}
                        </button>
                      ))}
                    </div>
                  )}
                </div>
              )}
              <button
                onClick={() => setStep("dentist")}
                className="mt-6 text-gray-500 hover:text-gray-700 text-sm"
              >
                ← Back to dentist
              </button>
            </div>
          )}

          {/* Step 4: Patient Info */}
          {step === "info" && (
            <div>
              <h2 className="text-2xl font-bold text-gray-900 mb-6">Your Information</h2>
              <div className="space-y-4 max-w-md">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Full Name *</label>
                  <input
                    type="text"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    className="w-full rounded-lg border border-gray-300 px-3 py-2.5 focus:outline-none focus:ring-2 focus:ring-teal-500"
                    placeholder="John Smith"
                    required
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Email *</label>
                  <input
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    className="w-full rounded-lg border border-gray-300 px-3 py-2.5 focus:outline-none focus:ring-2 focus:ring-teal-500"
                    placeholder="john@example.com"
                    required
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Phone *</label>
                  <input
                    type="tel"
                    value={phone}
                    onChange={(e) => setPhone(e.target.value)}
                    className="w-full rounded-lg border border-gray-300 px-3 py-2.5 focus:outline-none focus:ring-2 focus:ring-teal-500"
                    placeholder="(905) 555-0000"
                    required
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Notes (optional)</label>
                  <textarea
                    value={notes}
                    onChange={(e) => setNotes(e.target.value)}
                    rows={3}
                    className="w-full rounded-lg border border-gray-300 px-3 py-2.5 focus:outline-none focus:ring-2 focus:ring-teal-500 resize-none"
                    placeholder="Any special requests or concerns?"
                  />
                </div>
                <div className="flex gap-3 pt-2">
                  <button
                    onClick={() => setStep("datetime")}
                    className="px-6 py-2.5 rounded-lg border border-gray-300 text-gray-700 hover:bg-gray-50 font-medium"
                  >
                    Back
                  </button>
                  <button
                    onClick={() => {
                      if (!name || !email || !phone) {
                        alert("Please fill in all required fields.");
                        return;
                      }
                      setStep("confirm");
                    }}
                    className="px-6 py-2.5 rounded-lg bg-teal-600 text-white hover:bg-teal-700 font-medium"
                  >
                    Review Booking
                  </button>
                </div>
              </div>
            </div>
          )}

          {/* Step 5: Confirm */}
          {step === "confirm" && (
            <div>
              <h2 className="text-2xl font-bold text-gray-900 mb-6">Confirm Your Appointment</h2>
              <div className="bg-white rounded-2xl border border-gray-100 shadow-sm p-6 max-w-lg">
                <div className="space-y-3">
                  <div className="flex justify-between py-2 border-b border-gray-100">
                    <span className="text-gray-500">Service</span>
                    <span className="font-semibold text-gray-900">{service}</span>
                  </div>
                  <div className="flex justify-between py-2 border-b border-gray-100">
                    <span className="text-gray-500">Dentist</span>
                    <span className="font-semibold text-gray-900">{dentist}</span>
                  </div>
                  <div className="flex justify-between py-2 border-b border-gray-100">
                    <span className="text-gray-500">Date</span>
                    <span className="font-semibold text-gray-900">
                      {new Date(date + "T12:00:00").toLocaleDateString("en-CA", {
                        weekday: "long",
                        year: "numeric",
                        month: "long",
                        day: "numeric",
                      })}
                    </span>
                  </div>
                  <div className="flex justify-between py-2 border-b border-gray-100">
                    <span className="text-gray-500">Time</span>
                    <span className="font-semibold text-gray-900">{formatTime(time)}</span>
                  </div>
                  <div className="flex justify-between py-2 border-b border-gray-100">
                    <span className="text-gray-500">Patient</span>
                    <span className="font-semibold text-gray-900">{name}</span>
                  </div>
                  <div className="flex justify-between py-2 border-b border-gray-100">
                    <span className="text-gray-500">Email</span>
                    <span className="font-semibold text-gray-900">{email}</span>
                  </div>
                  <div className="flex justify-between py-2">
                    <span className="text-gray-500">Phone</span>
                    <span className="font-semibold text-gray-900">{phone}</span>
                  </div>
                  {notes && (
                    <div className="pt-2 border-t border-gray-100">
                      <span className="text-gray-500 text-sm">Notes: </span>
                      <span className="text-gray-700 text-sm">{notes}</span>
                    </div>
                  )}
                </div>
                <div className="flex gap-3 mt-6">
                  <button
                    onClick={() => setStep("info")}
                    className="flex-1 px-6 py-3 rounded-lg border border-gray-300 text-gray-700 hover:bg-gray-50 font-medium"
                  >
                    Back
                  </button>
                  <button
                    onClick={handleSubmit}
                    disabled={loading}
                    className="flex-1 px-6 py-3 rounded-lg bg-teal-600 text-white hover:bg-teal-700 font-medium disabled:bg-teal-400"
                  >
                    {loading ? "Booking..." : "Confirm Booking"}
                  </button>
                </div>
              </div>
            </div>
          )}

          {/* Success */}
          {step === "success" && (
            <div className="text-center py-8">
              <div className="w-20 h-20 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-6">
                <span className="text-4xl">✅</span>
              </div>
              <h2 className="text-3xl font-bold text-gray-900 mb-3">Appointment Booked!</h2>
              <p className="text-gray-600 text-lg mb-2">
                Your appointment has been confirmed.
              </p>
              <p className="text-gray-500 text-sm mb-8">
                Appointment #{appointmentId} · A confirmation email will be sent to {email}
              </p>
              <div className="bg-white rounded-2xl border border-gray-100 shadow-sm p-6 max-w-sm mx-auto mb-8">
                <div className="space-y-2 text-sm">
                  <p><span className="text-gray-500">Service:</span> <strong>{service}</strong></p>
                  <p><span className="text-gray-500">Dentist:</span> <strong>{dentist}</strong></p>
                  <p>
                    <span className="text-gray-500">When:</span>{" "}
                    <strong>
                      {new Date(date + "T12:00:00").toLocaleDateString("en-CA", {
                        weekday: "short",
                        month: "long",
                        day: "numeric",
                      })}{" "}
                      at {formatTime(time)}
                    </strong>
                  </p>
                </div>
              </div>
              <a
                href="/"
                className="inline-flex items-center justify-center bg-teal-600 text-white px-8 py-3 rounded-xl font-semibold hover:bg-teal-700 transition-colors"
              >
                Back to Home
              </a>
            </div>
          )}
        </div>
      </section>
    </>
  );
}
