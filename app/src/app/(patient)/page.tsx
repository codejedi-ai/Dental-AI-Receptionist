import Link from "next/link";
import { clinicConfig } from "@/lib/clinic-config";

export default function HomePage() {
  return (
    <>
      {/* Hero Section */}
      <section className="relative bg-gradient-to-br from-teal-600 via-teal-700 to-teal-800 text-white">
        <div className="absolute inset-0 bg-[url('data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNjAiIGhlaWdodD0iNjAiIHZpZXdCb3g9IjAgMCA2MCA2MCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48ZyBmaWxsPSJub25lIiBmaWxsLXJ1bGU9ImV2ZW5vZGQiPjxnIGZpbGw9IiNmZmZmZmYiIGZpbGwtb3BhY2l0eT0iMC4wNCI+PGNpcmNsZSBjeD0iMzAiIGN5PSIzMCIgcj0iMiIvPjwvZz48L2c+PC9zdmc+')] opacity-50" />
        <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-24 sm:py-32">
          <div className="max-w-3xl">
            <h1 className="text-4xl sm:text-5xl lg:text-6xl font-bold leading-tight mb-6">
              Your Smile Deserves the{" "}
              <span className="text-teal-200">Best Care</span>
            </h1>
            <p className="text-lg sm:text-xl text-teal-100 mb-8 leading-relaxed">
              Welcome to Smile Dental Clinic in Newmarket, Ontario. We provide
              comprehensive dental care for the whole family in a warm,
              comfortable environment.
            </p>
            <div className="flex flex-col sm:flex-row gap-4">
              <Link
                href="/book"
                className="inline-flex items-center justify-center bg-white text-teal-700 px-8 py-3.5 rounded-xl font-semibold text-lg hover:bg-teal-50 transition-colors shadow-lg"
              >
                Book Appointment
              </Link>
              <Link
                href="/services"
                className="inline-flex items-center justify-center border-2 border-white/30 text-white px-8 py-3.5 rounded-xl font-semibold text-lg hover:bg-white/10 transition-colors"
              >
                Our Services
              </Link>
            </div>
            <div className="mt-8 flex items-center space-x-6 text-teal-200 text-sm">
              <span className="flex items-center space-x-1">
                <span>📍</span>
                <span>Newmarket, ON</span>
              </span>
              <span className="flex items-center space-x-1">
                <span>📞</span>
                <span>{clinicConfig.phone}</span>
              </span>
              <span className="flex items-center space-x-1">
                <span>🕐</span>
                <span>Mon–Sat</span>
              </span>
            </div>
          </div>
        </div>
      </section>

      {/* Services Overview */}
      <section className="py-20 bg-gray-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-12">
            <h2 className="text-3xl sm:text-4xl font-bold text-gray-900 mb-4">
              Our Dental Services
            </h2>
            <p className="text-lg text-gray-600 max-w-2xl mx-auto">
              From routine checkups to advanced procedures, we offer a full
              range of dental services for patients of all ages.
            </p>
          </div>
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-4 sm:gap-6">
            {clinicConfig.services.map((service) => (
              <Link
                key={service.name}
                href="/services"
                className="bg-white rounded-xl p-5 text-center hover:shadow-lg transition-shadow border border-gray-100 group"
              >
                <span className="text-4xl block mb-3 group-hover:scale-110 transition-transform">
                  {service.icon}
                </span>
                <h3 className="font-semibold text-gray-900 text-sm">
                  {service.name}
                </h3>
              </Link>
            ))}
          </div>
          <div className="text-center mt-8">
            <Link
              href="/services"
              className="text-teal-600 font-semibold hover:text-teal-700 transition-colors"
            >
              View All Services →
            </Link>
          </div>
        </div>
      </section>

      {/* Meet Our Dentists */}
      <section className="py-20">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-12">
            <h2 className="text-3xl sm:text-4xl font-bold text-gray-900 mb-4">
              Meet Our Dentists
            </h2>
            <p className="text-lg text-gray-600 max-w-2xl mx-auto">
              Our experienced team is dedicated to providing you with the
              highest quality dental care.
            </p>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            {clinicConfig.dentists.map((dentist) => (
              <div
                key={dentist.id}
                className="bg-white rounded-2xl p-6 border border-gray-100 shadow-sm hover:shadow-md transition-shadow"
              >
                <div className="w-20 h-20 bg-teal-100 rounded-full flex items-center justify-center mx-auto mb-4">
                  <span className="text-3xl">👨‍⚕️</span>
                </div>
                <h3 className="text-xl font-bold text-gray-900 text-center mb-1">
                  {dentist.name}
                </h3>
                <p className="text-teal-600 text-sm font-medium text-center mb-3">
                  {dentist.specialty}
                </p>
                <p className="text-gray-600 text-sm text-center leading-relaxed">
                  {dentist.bio}
                </p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Why Choose Us */}
      <section className="py-20 bg-teal-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-12">
            <h2 className="text-3xl sm:text-4xl font-bold text-gray-900 mb-4">
              Why Choose Smile Dental?
            </h2>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
            {[
              { icon: "🏆", title: "Experienced Team", desc: "Over 40 years of combined dental experience" },
              { icon: "🔬", title: "Modern Technology", desc: "State-of-the-art equipment and techniques" },
              { icon: "💖", title: "Patient Comfort", desc: "Gentle care in a warm, relaxing environment" },
              { icon: "🕐", title: "Flexible Hours", desc: "Evening and Saturday appointments available" },
            ].map((item) => (
              <div key={item.title} className="bg-white rounded-xl p-6 text-center shadow-sm">
                <span className="text-4xl block mb-3">{item.icon}</span>
                <h3 className="font-bold text-gray-900 mb-2">{item.title}</h3>
                <p className="text-gray-600 text-sm">{item.desc}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="py-16 bg-teal-700 text-white">
        <div className="max-w-4xl mx-auto px-4 text-center">
          <h2 className="text-3xl sm:text-4xl font-bold mb-4">
            Ready for a Healthier Smile?
          </h2>
          <p className="text-teal-100 text-lg mb-8">
            Book your appointment today or talk to Lisa, our AI receptionist,
            anytime.
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Link
              href="/book"
              className="inline-flex items-center justify-center bg-white text-teal-700 px-8 py-3.5 rounded-xl font-semibold text-lg hover:bg-teal-50 transition-colors"
            >
              Book Online
            </Link>
            <a
              href={`tel:${clinicConfig.phone.replace(/[^0-9]/g, "")}`}
              className="inline-flex items-center justify-center border-2 border-white/30 text-white px-8 py-3.5 rounded-xl font-semibold text-lg hover:bg-white/10 transition-colors"
            >
              📞 Call {clinicConfig.phone}
            </a>
          </div>
        </div>
      </section>
    </>
  );
}
