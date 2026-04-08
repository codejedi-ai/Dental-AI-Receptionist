import { clinicConfig } from "@/lib/clinic-config";

export default function AboutPage() {
  return (
    <>
      <section className="bg-gradient-to-br from-teal-600 to-teal-800 text-white py-16">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <h1 className="text-4xl sm:text-5xl font-bold mb-4">About Us</h1>
          <p className="text-teal-100 text-lg max-w-2xl">Providing exceptional dental care to Newmarket and the surrounding community since 2010.</p>
        </div>
      </section>

      <section className="py-16">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="max-w-3xl mx-auto text-center mb-16">
            <h2 className="text-3xl font-bold text-gray-900 mb-6">Our Story</h2>
            <p className="text-gray-600 text-lg leading-relaxed mb-4">Smile Dental Clinic was founded with a simple mission: to provide high-quality, compassionate dental care in a comfortable environment. Located in the heart of Newmarket, Ontario, we've been serving families for over a decade.</p>
            <p className="text-gray-600 text-lg leading-relaxed">Our team combines expertise with a gentle touch, ensuring every visit is a positive experience.</p>
          </div>

          <h2 className="text-3xl font-bold text-gray-900 text-center mb-8">Our Team</h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8 mb-16">
            {clinicConfig.dentists.map((d) => (
              <div key={d.id} className="bg-white rounded-2xl border border-gray-100 shadow-sm p-8 text-center">
                <div className="w-24 h-24 bg-teal-100 rounded-full flex items-center justify-center mx-auto mb-4"><span className="text-4xl">👨‍⚕️</span></div>
                <h3 className="text-xl font-bold text-gray-900 mb-1">{d.name}</h3>
                <p className="text-teal-600 font-medium text-sm mb-4">{d.specialty}</p>
                <p className="text-gray-600 text-sm leading-relaxed">{d.bio}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="py-16 bg-gray-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-12">
            <div>
              <h2 className="text-3xl font-bold text-gray-900 mb-6">Contact Us</h2>
              <div className="space-y-4">
                <div className="flex items-start space-x-3"><span className="text-2xl mt-1">📍</span><div><h3 className="font-semibold text-gray-900">Address</h3><p className="text-gray-600">{clinicConfig.address}</p></div></div>
                <div className="flex items-start space-x-3"><span className="text-2xl mt-1">📞</span><div><h3 className="font-semibold text-gray-900">Phone</h3><p className="text-gray-600">{clinicConfig.phone}</p></div></div>
                <div className="flex items-start space-x-3"><span className="text-2xl mt-1">✉️</span><div><h3 className="font-semibold text-gray-900">Email</h3><p className="text-gray-600">{clinicConfig.email}</p></div></div>
              </div>
              <h3 className="text-xl font-bold text-gray-900 mt-8 mb-4">Office Hours</h3>
              <div className="bg-white rounded-xl border border-gray-100 overflow-hidden">
                {clinicConfig.hoursDetailed.map((h, i) => (
                  <div key={h.day} className={`flex justify-between px-4 py-3 ${i % 2 === 0 ? "bg-white" : "bg-gray-50"}`}>
                    <span className="font-medium text-gray-900">{h.day}</span>
                    <span className={h.hours === "Closed" ? "text-red-500" : "text-gray-600"}>{h.hours}</span>
                  </div>
                ))}
              </div>
            </div>
            <div>
              <h2 className="text-3xl font-bold text-gray-900 mb-6">Send Us a Message</h2>
              <form className="space-y-4" onSubmit={(e) => e.preventDefault()}>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div><label className="block text-sm font-medium text-gray-700 mb-1">First Name</label><input type="text" className="w-full rounded-lg border border-gray-300 px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-teal-500" placeholder="John" /></div>
                  <div><label className="block text-sm font-medium text-gray-700 mb-1">Last Name</label><input type="text" className="w-full rounded-lg border border-gray-300 px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-teal-500" placeholder="Doe" /></div>
                </div>
                <div><label className="block text-sm font-medium text-gray-700 mb-1">Email</label><input type="email" className="w-full rounded-lg border border-gray-300 px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-teal-500" placeholder="john@example.com" /></div>
                <div><label className="block text-sm font-medium text-gray-700 mb-1">Message</label><textarea rows={4} className="w-full rounded-lg border border-gray-300 px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-teal-500 resize-none" placeholder="How can we help?" /></div>
                <button type="submit" className="w-full bg-teal-600 text-white py-3 rounded-lg font-semibold hover:bg-teal-700 transition-colors">Send Message</button>
              </form>
              <div className="mt-8 bg-gray-200 rounded-xl h-48 flex items-center justify-center">
                <div className="text-center text-gray-500"><span className="text-4xl block mb-2">🗺️</span><p className="text-sm font-medium">123 Main Street, Newmarket, ON</p></div>
              </div>
            </div>
          </div>
        </div>
      </section>
    </>
  );
}
