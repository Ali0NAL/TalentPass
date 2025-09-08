import { useState } from "react";
import { api } from "../api";
import { setToken } from "../auth";
import { useNavigate } from "react-router-dom";

export default function Login() {
  const [email, setEmail] = useState("demo@example.com");
  const [password, setPassword] = useState("S3cure!");
  const [err, setErr] = useState("");
  const nav = useNavigate();

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setErr("");
    try {
      const { data } = await api.post("/v1/auth/login", { email, password });
      setToken(data.access_token);
      nav("/jobs");
    } catch (e: any) {
      setErr(e.response?.data?.error || "Login failed");
    }
  }

  return (
    <div className="min-h-screen grid place-items-center p-6">
      <form onSubmit={onSubmit} className="w-full max-w-sm space-y-3 p-6 rounded-2xl shadow">
        <h1 className="text-xl font-semibold">Giriş</h1>
        {err && <div className="text-red-600 text-sm">{err}</div>}
        <input className="w-full border p-2 rounded" value={email} onChange={e=>setEmail(e.target.value)} placeholder="Email" />
        <input className="w-full border p-2 rounded" type="password" value={password} onChange={e=>setPassword(e.target.value)} placeholder="Şifre" />
        <button className="w-full bg-black text-white py-2 rounded">Giriş Yap</button>
      </form>
    </div>
  );
}
