import { Routes, Route } from "react-router-dom";
import PatientLayout from "@/components/PatientLayout";
import AdminLayout from "@/components/AdminLayout";
import HomePage from "@/pages/HomePage";
import ServicesPage from "@/pages/ServicesPage";
import BookPage from "@/pages/BookPage";
import AboutPage from "@/pages/AboutPage";
import DashboardPage from "@/pages/admin/DashboardPage";
import AppointmentsPage from "@/pages/admin/AppointmentsPage";

export default function App() {
  return (
    <Routes>
      <Route element={<PatientLayout />}>
        <Route path="/" element={<HomePage />} />
        <Route path="/services" element={<ServicesPage />} />
        <Route path="/book" element={<BookPage />} />
        <Route path="/about" element={<AboutPage />} />
      </Route>
      <Route element={<AdminLayout />}>
        <Route path="/admin" element={<DashboardPage />} />
        <Route path="/admin/appointments" element={<AppointmentsPage />} />
      </Route>
    </Routes>
  );
}
