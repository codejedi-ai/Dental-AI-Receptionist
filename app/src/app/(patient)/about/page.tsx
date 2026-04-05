import { clinicConfig } from "@/lib/clinic-config";

export const metadata = {
  title: "About Us — Smile Dental Clinic",
  description: "Learn about Smile Dental Clinic in Newmarket, Ontario. Meet our team and discover our approach to dental care.",
};

export default function AboutPage() {
  return (
    <>
      {/* Header */}
      <section className="bg-gradient-to-br from-teal-600 to-teal-800 text-white py-16">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <h1 className="text-4xl sm:text-5xl font-bold mb-4">About Us</h1>
          <p className="text-teal-100 text-lg max-w-2xl">
            Providing exceptional dental care to Newmarket and the surrounding community since 2010.
          </p>
        </div>
      </section>

      {/* Our Story */}
      <section className="py-16">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="max-w-3xl mx-auto text-center mb-16">
            <h2 className="text-3xl font-bold text-gray-900 mb-6">Our Story</h2>
            <p className="text-gray-600 text-lg leading-relaxed mb-4">
              Smile Dental Clinic was founded with a simple mission: to provide high-quality, compassionate dental care in a comfortable environment. Located in the heart of Newmarket, Ontario, we&apos;ve been serving families for over a decade.
            </p>
            <p className="text-gray-600 text-lg leading-relaxed">
              Our team of skilled professionals combines expertise with a gentle touch, ensuring that every visit is a positive experience. We invest in the latest technology and continuing education to bring you the best in modern dentistry.
            </p>
          </div>

          {/* Team */}
          <h2 className="text-3xl font-bold text-gray-900 text-center mb-8">Our Team</h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8 mb-16">
            {clinicConfig.dentists.map((dentist) => (
              <div
                key={dentist.id}
                className="bg-white rounded-2xl border border-gray-100 shadow-sm p-8 text-center"
              >
                <div className="w-24 h-24 bg-teal-100 rounded-full flex items-center justify-center mx-auto mb-4">
                  <span className="text-4xl">👨‍⚕️</span>
                </div>
                <h3 className="text-xl font-bold text-gray-900 mb-1">{dentist.name}</h3>
                <p className="text-teal-600 font-medium text-sm mb-4">{dentist.specialty}</p>
                <p className="text-gray-600 text-sm leading-relaxed">{dentist.bio}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Clinic Info */}
      <section className="py-16 bg-gray-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-12">
            {/* Contact Info */}
            <div>
              <h2 className="text-3xl font-bold text-gray-900 mb-6">Contact Us</h2>
              <div className="space-y-4">
                <div className="flex items-start space-x-3">
                  <span className="text-2xl mt-1">📍</span>
                  <div>
                    <h3 className="font-semibold text-gray-900">Address</h3>
                    <p className="text-gray-600">{clinicConfig.address}</p>
                  </div>
                </div>
                <div className="flex items-start space-x-3">
                  <span className="text-2xl mt-1">📞</span>
                  <div>
                    <h3 className="font-semibold text-gray-900">Phone</h3>
                    <p className="text-gray-600">{clinicConfig.phone}</p>
                  </div>
                </div>
                <div className="flex items-start space-x-3">
                  <span className="text-2xl mt-1">✉️</span>
                  <div>
                    <h3 className="font-semibold text-gray-900">Email</h3>
                    <p className="text-gray-600">{clinicConfig.email}</p>
                  </div>
                </div>
              </div>

              {/* Hours */}
              <h3 className="text-xl font-bold text-gray-900 mt-8 mb-4">Office Hours</h3>
              <div className="bg-white rounded-xl border border-gray-100 overflow-hidden">
                {clinicConfig.hoursDetailed.map((h, i) => (
                  <div
                    key={h.day}
                    className={`flex justify-between px-4 py-3 ${
                      i % 2 === 0 ? "bg-white" : "bg-gray-50"
                    }`}
                  >
                    <span className="font-medium text-gray-900">{h.day}</span>
                    <span className={`${h.hours === "Closed" ? "text-red-500" : "text-gray-600"}`}>
                      {h.hours}
                    </span>
                  </div>
                ))}
              </div>
            </div>

            {/* Contact Form */}
            <div>
              <h2 className="text-3xl font-bold text-gray-900 mb-6">Send Us a Message</h2>
              <form className="space-y-4">
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">First Name</label>
                    <input
                      type="text"
                      className="w-full rounded-lg border border-gray-300 px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-teal-500"
                      placeholder="John"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Last Name</label>
                    <input
                      type="text"
                      className="w-full rounded-lg border border-gray-300 px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-teal-500"
                      placeholder="Doe"
                    />
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Email</label>
                  <input
                    type="email"
                    className="w-full rounded-lg border border-gray-300 px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-teal-500"
                    placeholder="john@example.com"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Phone</label>
                  <input
                    type="tel"
                    className="w-full rounded-lg border border-gray-300 px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-teal-500"
                    placeholder="(905) 555-0000"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Message</label>
                  <textarea
                    rows={4}
                    className="w-full rounded-lg border border-gray-300 px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-teal-500 resize-none"
                    placeholder="How can we help you?"
                  />
                </div>
                <button
                  type="submit"
                  className="w-full bg-teal-600 text-white py-3 rounded-lg font-semibold hover:bg-teal-700 transition-colors"
                >
                  Send Message
                </button>
              </form>

              {/* Map Placeholder */}
              <div className="mt-8 bg-gray-200 rounded-xl h-48 flex items-center justify-center">
                <div className="text-center text-gray-500">
                  <span className="text-4xl block mb-2">🗺️</span>
                  <p className="text-sm font-medium">123 Main Street, Newmarket, ON</p>
                  <p className="text-xs">Map integration available</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>
    </>
  );
}
