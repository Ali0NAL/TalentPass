import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { api } from "../api";

type Event = { id:number; type:string; payload_json:any; created_at:string };

export default function ApplicationEvents() {
  const { id } = useParams();
  const [events, setEvents] = useState<Event[]>([]);
  const [msg, setMsg] = useState("");

  const load = () => api.get(`/v1/applications/${id}/events?limit=50`).then(r => setEvents(r.data.items));
  useEffect(() => { load(); }, [id]);

  async function addNote() {
    await api.post(`/v1/applications/${id}/events`, {
      type: "note",
      payload_json: { message: msg }
    });
    setMsg("");
    load();
  }

  return (
    <div className="p-6 max-w-2xl mx-auto">
      <h1 className="text-xl font-semibold mb-4">Başvuru #{id} – Olaylar</h1>

      <div className="flex gap-2 mb-4">
        <input className="flex-1 border p-2 rounded" placeholder="Not…" value={msg} onChange={e=>setMsg(e.target.value)} />
        <button className="bg-black text-white px-4 rounded" onClick={addNote}>Ekle</button>
      </div>

      <div className="space-y-2">
        {events.map(ev => (
          <div key={ev.id} className="p-3 border rounded-xl">
            <div className="text-sm text-gray-600">{new Date(ev.created_at).toLocaleString()} • {ev.type}</div>
            {ev.payload_json?.message && <div className="mt-1">{ev.payload_json.message}</div>}
          </div>
        ))}
      </div>
    </div>
  );
}
