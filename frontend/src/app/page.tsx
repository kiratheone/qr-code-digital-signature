export default function Home() {
  return (
    <main className="container mx-auto px-4 py-8">
      <div className="text-center">
        <h1 className="text-4xl font-bold text-gray-900 mb-4">
          Digital Signature System
        </h1>
        <p className="text-xl text-gray-600 mb-8">
          QR Code-based digital signature system for PDF documents
        </p>
        <div className="space-y-4">
          <div className="bg-white p-6 rounded-lg shadow-md">
            <h2 className="text-2xl font-semibold mb-4">Features</h2>
            <ul className="text-left space-y-2">
              <li>✅ Digital document signing with SHA-256 hash</li>
              <li>✅ QR Code generation and PDF injection</li>
              <li>✅ Document management and verification</li>
              <li>✅ Secure authentication system</li>
            </ul>
          </div>
        </div>
      </div>
    </main>
  )
}