import { useEffect, useState } from "react";
import { api } from "../api";
import { Link } from "react-router-dom";

type Job = { id:number; title:string; company:string; location:string; updated_at?:string };

export default function Jobs() {
  const [items, setItems] = useState<Job[]>([]);
  useEffect(() => { api.get("/v1/jobs").then(r => setItems(r.data.items)); }, []);
  return (
    <div className="p-6 max-w-3xl mx-auto">
      <h1 className="text-2xl font-semibold mb-4">İşler</h1>
      <div className="space-y-2">
        {items.map(j => (
          <Link key={j.id} to={`/applications/${j.id}`} className="block p-4 rounded-xl border hover:bg-gray-50">
            <div className="font-medium">{j.title}</div>
            <div className="text-sm text-gray-600">{j.company} • {j.location}</div>
          </Link>
        ))}
      </div>
    </div>
  );
}
