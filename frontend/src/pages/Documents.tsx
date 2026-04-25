import { useEffect, useRef, useState } from "react";
import api from "../api/client";
import { Upload, Trash2, FileText, Loader2 } from "lucide-react";

type Doc = {
  id: string; title: string; source_type: string;
  status: string; created_at: string;
};

const statusColor: Record<string, string> = {
  ready:      "bg-green-900 text-green-300",
  processing: "bg-yellow-900 text-yellow-300",
  pending:    "bg-gray-700 text-gray-300",
  failed:     "bg-red-900 text-red-300",
};

export default function Documents() {
  const [docs,     setDocs]     = useState<Doc[]>([]);
  const [loading,  setLoading]  = useState(false);
  const [uploading,setUploading]= useState(false);
  const fileRef = useRef<HTMLInputElement>(null);

  const load = async () => {
    setLoading(true);
    const { data } = await api.get("/documents");
    setDocs(data.documents);
    setLoading(false);
  };

  useEffect(() => { load(); }, []);

  const upload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    setUploading(true);
    const form = new FormData();
    form.append("file", file);
    await api.post("/documents", form, {
      headers: { "Content-Type": "multipart/form-data" }
    });
    await load();
    setUploading(false);
    if (fileRef.current) fileRef.current.value = "";
  };

  const remove = async (id: string) => {
    await api.delete(`/documents/${id}`);
    setDocs(d => d.filter(x => x.id !== id));
  };

  return (
    <div className="h-full flex flex-col p-6">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-lg font-semibold">Documents</h2>
        <label className="flex items-center gap-2 bg-indigo-600 hover:bg-indigo-500 text-sm px-4 py-2 rounded-lg cursor-pointer transition-colors">
          {uploading
            ? <><Loader2 size={15} className="animate-spin" /> Uploading…</>
            : <><Upload size={15} /> Upload</>}
          <input ref={fileRef} type="file" accept=".pdf,.txt,.docx"
            className="hidden" onChange={upload} disabled={uploading} />
        </label>
      </div>

      {loading ? (
        <div className="flex-1 flex items-center justify-center text-gray-500">
          <Loader2 size={24} className="animate-spin" />
        </div>
      ) : docs.length === 0 ? (
        <div className="flex-1 flex flex-col items-center justify-center text-gray-500 gap-3">
          <FileText size={40} strokeWidth={1} />
          <p className="text-sm">No documents yet — upload a PDF, DOCX or TXT</p>
        </div>
      ) : (
        <div className="space-y-2 overflow-y-auto">
          {docs.map(doc => (
            <div key={doc.id}
              className="flex items-center justify-between bg-gray-900 border border-gray-800 rounded-lg px-4 py-3">
              <div className="flex items-center gap-3 min-w-0">
                <FileText size={16} className="text-gray-400 shrink-0" />
                <div className="min-w-0">
                  <p className="text-sm truncate">{doc.title}</p>
                  <p className="text-xs text-gray-500">
                    {doc.source_type.toUpperCase()} · {new Date(doc.created_at).toLocaleDateString()}
                  </p>
                </div>
              </div>
              <div className="flex items-center gap-3 shrink-0 ml-4">
                <span className={`text-xs px-2 py-0.5 rounded-full ${statusColor[doc.status] ?? "bg-gray-700 text-gray-300"}`}>
                  {doc.status}
                </span>
                <button onClick={() => remove(doc.id)}
                  className="text-gray-500 hover:text-red-400 transition-colors">
                  <Trash2 size={15} />
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}