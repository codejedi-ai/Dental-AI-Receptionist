import { Link } from "react-router-dom";
import { clinicConfig } from "@/lib/clinic-config";

export default function HomePage() {
  return (
    <>
      <section className="relative bg-gradient-to-br from-teal-600 via-teal-700 to-teal-800 text-white">
        <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-24 sm:py-32">
          <div className="max-w-3xl">
            <h1 className="text-4xl sm:text-5xl lg:text-6xl font-bold leading-tight mb-6">Your Smile Deserves the <span className="text-teal-200">Best Care</span></h1>
            <p className="text-lg sm:text-xl text-teal-100 mb-8 leading-relaxed">Welcome to Smile Dental Clinic in Newmarket, Ontario. We provide comprehensive dental care for the whole family in a warm, comfortable environment.</p>
            <div className="flex flex-col sm:flex-row gap-4">
              <Link to="/book" className="inline-flex items-center justify-center bg-white text-teal-700 px-8 py-3.5 rounded-xl font-semibold text-lg hover:bg-teal-50 transition-colors shadow-lg">Book Appointment</Link>
              <Link to="/services" className="inline-flex items-center justify-center border-2 border-white/30 text-white px-8 py-3.5 rounded-xl font-semibold text-lg hover:bg-white/10 transition-colors">Our Services</Link>
            </div>
            <div className="mt-8 flex items-center space-x-6 text-teal-200 text-sm">
              <span>📍 Newmarket, ON</span>
              <span>📞 {clinicConfig.phone}</span>
              <span>🕐 Mon–Sat</span>
            </div>
          </div>
        </div>
      </section>

      <section className="py-20 bg-gray-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-12">
            <h2 className="text-3xl sm:text-4xl font-bold text-gray-900 mb-4">Our Dental Services</h2>
            <p className="text-lg text-gray-600 max-w-2xl mx-auto">From routine checkups to advanced procedures, we offer a full range of dental services for patients of all ages.</p>
          </div>
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-4 sm:gap-6">
            {clinicConfig.services.map((s) => (
              <Link key={s.name} to="/services" className="bg-white rounded-xl p-5 text-center hover:shadow-lg transition-shadow border border-gray-100 group">
                <span className="text-4xl block mb-3 group-hover:scale-110 transition-transform">{s.icon}</span>
                <h3 className="font-semibold text-gray-900 text-sm">{s.name}</h3>
              </Link>
            ))}
          </div>
        </div>
      </section>

      <section className="py-20">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-12">
            <h2 className="text-3xl sm:text-4xl font-bold text-gray-900 mb-4">Meet Our Dentists</h2>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            {clinicConfig.dentists.map((d) => (
              <div key={d.id} className="bg-white rounded-2xl p-6 border border-gray-100 shadow-sm hover:shadow-md transition-shadow">
                <div className="w-20 h-20 bg-teal-100 rounded-full flex items-center justify-center mx-auto mb-4"><span className="text-3xl">👨‍⚕️</span></div>
                <h3 className="text-xl font-bold text-gray-900 text-center mb-1">{d.name}</h3>
                <p className="text-teal-600 text-sm font-medium text-center mb-3">{d.specialty}</p>
                <p className="text-gray-600 text-sm text-center leading-relaxed">{d.bio}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="py-20 bg-teal-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-12"><h2 className="text-3xl sm:text-4xl font-bold text-gray-900">Why Choose Smile Dental?</h2></div>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
            {[
              { icon: "🏆", title: "Experienced Team", desc: "Over 40 years of combined dental experience" },
              { icon: "🔬", title: "Modern Technology", desc: "State-of-the-art equipment and techniques" },
              { icon: "💖", title: "Patient Comfort", desc: "Gentle care in a warm, relaxing environment" },
              { icon: "🕐", title: "Flexible Hours", desc: "Evening and Saturday appointments available" },
            ].map((i) => (
              <div key={i.title} className="bg-white rounded-xl p-6 text-center shadow-sm">
                <span className="text-4xl block mb-3">{i.icon}</span>
                <h3 className="font-bold text-gray-900 mb-2">{i.title}</h3>
                <p className="text-gray-600 text-sm">{i.desc}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="py-16 bg-teal-700 text-white">
        <div className="max-w-4xl mx-auto px-4 text-center">
          <h2 className="text-3xl sm:text-4xl font-bold mb-4">Ready for a Healthier Smile?</h2>
          <p className="text-teal-100 text-lg mb-8">Book your appointment today or talk to Lisa, our AI receptionist, anytime.</p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Link to="/book" className="inline-flex items-center justify-center bg-white text-teal-700 px-8 py-3.5 rounded-xl font-semibold text-lg hover:bg-teal-50 transition-colors">Book Online</Link>
            <a href={`tel:${clinicConfig.phone.replace(/[^0-9]/g, "")}`} className="inline-flex items-center justify-center border-2 border-white/30 text-white px-8 py-3.5 rounded-xl font-semibold text-lg hover:bg-white/10 transition-colors">📞 Call {clinicConfig.phone}</a>
          </div>
        </div>
      </section>
    </>
  );
}
