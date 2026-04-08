import { Link } from "react-router-dom";
import { useState } from "react";

export default function Navbar() {
  const [mobileOpen, setMobileOpen] = useState(false);

  return (
    <nav className="sticky top-0 z-50 bg-white/95 backdrop-blur-sm border-b border-gray-100 shadow-sm">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between h-16 items-center">
          <Link to="/" className="flex items-center space-x-2">
            <span className="text-3xl">🦷</span>
            <span className="text-xl font-bold text-teal-700">Smile Dental</span>
          </Link>
          <div className="hidden md:flex items-center space-x-8">
            <Link to="/" className="text-gray-600 hover:text-teal-600 transition-colors font-medium">Home</Link>
            <Link to="/services" className="text-gray-600 hover:text-teal-600 transition-colors font-medium">Services</Link>
            <Link to="/about" className="text-gray-600 hover:text-teal-600 transition-colors font-medium">About</Link>
            <Link to="/book" className="bg-teal-600 text-white px-5 py-2.5 rounded-lg hover:bg-teal-700 transition-colors font-medium">Book Appointment</Link>
          </div>
          <button className="md:hidden p-2 rounded-lg hover:bg-gray-100" onClick={() => setMobileOpen(!mobileOpen)}>
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              {mobileOpen ? (
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              ) : (
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
              )}
            </svg>
          </button>
        </div>
        {mobileOpen && (
          <div className="md:hidden pb-4 space-y-2">
            <Link to="/" className="block px-3 py-2 rounded-lg hover:bg-gray-50 text-gray-600" onClick={() => setMobileOpen(false)}>Home</Link>
            <Link to="/services" className="block px-3 py-2 rounded-lg hover:bg-gray-50 text-gray-600" onClick={() => setMobileOpen(false)}>Services</Link>
            <Link to="/about" className="block px-3 py-2 rounded-lg hover:bg-gray-50 text-gray-600" onClick={() => setMobileOpen(false)}>About</Link>
            <Link to="/book" className="block px-3 py-2 rounded-lg bg-teal-600 text-white text-center" onClick={() => setMobileOpen(false)}>Book Appointment</Link>
          </div>
        )}
      </div>
    </nav>
  );
}
