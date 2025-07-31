export default function HomePage() {
  return (
    <div className="space-y-8">
      <div className="text-center">
        <h1 className="text-3xl font-bold text-gray-900 mb-4">
          Digital Signature System
        </h1>
        <p className="text-lg text-gray-600 max-w-2xl mx-auto">
          Secure QR Code-based digital signature system for PDF documents. 
          Sign documents with cryptographic signatures and verify authenticity through QR codes.
        </p>
      </div>
      
      <div className="grid md:grid-cols-2 gap-8">
        <div className="card">
          <h2 className="text-xl font-semibold mb-4">Sign Documents</h2>
          <p className="text-gray-600 mb-4">
            Upload PDF documents and create digital signatures with embedded QR codes for verification.
          </p>
          <button className="btn btn-primary">
            Get Started
          </button>
        </div>
        
        <div className="card">
          <h2 className="text-xl font-semibold mb-4">Verify Documents</h2>
          <p className="text-gray-600 mb-4">
            Scan QR codes or upload documents to verify their authenticity and integrity.
          </p>
          <button className="btn btn-secondary">
            Verify Document
          </button>
        </div>
      </div>
    </div>
  )
}