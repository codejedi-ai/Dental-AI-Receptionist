import Link from "next/link";
import { clinicConfig } from "@/lib/clinic-config";

export const metadata = {
  title: "Our Services — Smile Dental Clinic",
  description: "Comprehensive dental services including checkups, cleanings, whitening, implants, Invisalign, and emergency care in Newmarket, ON.",
};

export default function ServicesPage() {
  return (
    <>
      {/* Header */}
      <section className="bg-gradient-to-br from-teal-600 to-teal-800 text-white py-16">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <h1 className="text-4xl sm:text-5xl font-bold mb-4">Our Services</h1>
          <p className="text-teal-100 text-lg max-w-2xl">
            We offer a comprehensive range of dental services to keep your smile healthy and beautiful. From preventive care to advanced treatments.
          </p>
        </div>
      </section>

      {/* Services Grid */}
      <section className="py-16">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {clinicConfig.services.map((service) => (
              <div
                key={service.name}
                className="bg-white rounded-2xl border border-gray-100 shadow-sm hover:shadow-lg transition-all p-6 group"
              >
                <div className="flex items-start space-x-4">
                  <span className="text-5xl group-hover:scale-110 transition-transform flex-shrink-0">
                    {service.icon}
                  </span>
                  <div className="flex-1">
                    <h2 className="text-xl font-bold text-gray-900 mb-2">
                      {service.name}
                    </h2>
                    <p className="text-gray-600 text-sm leading-relaxed mb-3">
                      {service.description}
                    </p>
                    <p className="text-teal-600 font-semibold text-sm">
                      {service.price}
                    </p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Insurance Info */}
      <section className="py-16 bg-gray-50">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
          <h2 className="text-3xl font-bold text-gray-900 mb-4">Insurance & Payment</h2>
          <p className="text-gray-600 text-lg mb-6">
            We accept most major dental insurance plans and offer flexible payment options.
            Direct billing is available for your convenience.
          </p>
          <div className="flex flex-wrap justify-center gap-4 mb-8">
            {["Sun Life", "Manulife", "Great-West Life", "Blue Cross", "Green Shield", "Desjardins"].map((ins) => (
              <span
                key={ins}
                className="bg-white border border-gray-200 rounded-lg px-4 py-2 text-sm font-medium text-gray-700"
              >
                {ins}
              </span>
            ))}
          </div>
          <Link
            href="/book"
            className="inline-flex items-center justify-center bg-teal-600 text-white px-8 py-3 rounded-xl font-semibold hover:bg-teal-700 transition-colors"
          >
            Book Your Appointment
          </Link>
        </div>
      </section>
    </>
  );
}
