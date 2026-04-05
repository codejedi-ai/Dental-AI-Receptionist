import Link from "next/link";
import { clinicConfig } from "@/lib/clinic-config";

export default function Footer() {
  return (
    <footer className="bg-gray-900 text-gray-300">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-8">
          {/* Brand */}
          <div className="md:col-span-1">
            <div className="flex items-center space-x-2 mb-4">
              <span className="text-3xl">🦷</span>
              <span className="text-xl font-bold text-white">Smile Dental</span>
            </div>
            <p className="text-sm text-gray-400">
              Your trusted family dentist in Newmarket, Ontario. Quality care for beautiful smiles.
            </p>
          </div>

          {/* Quick Links */}
          <div>
            <h3 className="text-white font-semibold mb-4">Quick Links</h3>
            <ul className="space-y-2 text-sm">
              <li><Link href="/" className="hover:text-teal-400 transition-colors">Home</Link></li>
              <li><Link href="/services" className="hover:text-teal-400 transition-colors">Services</Link></li>
              <li><Link href="/about" className="hover:text-teal-400 transition-colors">About Us</Link></li>
              <li><Link href="/book" className="hover:text-teal-400 transition-colors">Book Appointment</Link></li>
            </ul>
          </div>

          {/* Hours */}
          <div>
            <h3 className="text-white font-semibold mb-4">Office Hours</h3>
            <ul className="space-y-2 text-sm">
              <li>Mon – Fri: {clinicConfig.hours.weekdays}</li>
              <li>Saturday: {clinicConfig.hours.saturday}</li>
              <li>Sunday: {clinicConfig.hours.sunday}</li>
            </ul>
          </div>

          {/* Contact */}
          <div>
            <h3 className="text-white font-semibold mb-4">Contact</h3>
            <ul className="space-y-2 text-sm">
              <li>📍 {clinicConfig.address}</li>
              <li>📞 {clinicConfig.phone}</li>
              <li>✉️ {clinicConfig.email}</li>
            </ul>
          </div>
        </div>

        <div className="border-t border-gray-800 mt-8 pt-8 text-center text-sm text-gray-500">
          <p>© {new Date().getFullYear()} {clinicConfig.name}. All rights reserved.</p>
        </div>
      </div>
    </footer>
  );
}
