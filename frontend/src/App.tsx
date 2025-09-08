import { useState } from 'react';
import { login, listApplications } from './api';

export default function App() {
  const [email, setEmail] = useState('demo@example.com');
  const [password, setPassword] = useState('S3cure!');
  const [token, setToken] = useState<string | null>(localStorage.getItem('tp_token'));
  const [msg, setMsg] = useState('');
  const [apps, setApps] = useState<Array<{ id: number; title: string; company: string; location: string }>>([]);

  async function onLogin(e: React.FormEvent) {
    e.preventDefault();
    setMsg('');
    try {
      const data = await login(email, password);
      localStorage.setItem('tp_token', data.access_token);
      setToken(data.access_token);
      setMsg(`Giriş OK: ${data.user.email}`);
    } catch (err: any) {
      setMsg(err.message ?? 'Login failed');
    }
  }

  async function loadApps() {
    if (!token) return setMsg('Önce login ol');
    try {
      const data = await listApplications(token);
      setApps(data.items ?? []);
      setMsg(`Toplam ${data.items?.length ?? 0} ilan`);
    } catch (err: any) {
      setMsg(err.message ?? 'Listeleme hatası');
    }
  }

  function logout() {
    localStorage.removeItem('tp_token');
    setToken(null);
    setApps([]);
    setMsg('Çıkış yapıldı');
  }

  return (
    <div style={{ padding: 24, fontFamily: 'system-ui', maxWidth: 720, margin: '0 auto' }}>
      <h1>TalentPass</h1>

      {!token ? (
        <form onSubmit={onLogin} style={{ display: 'grid', gap: 8, maxWidth: 360 }}>
          <input value={email} onChange={(e) => setEmail(e.target.value)} placeholder="email" />
          <input value={password} onChange={(e) => setPassword(e.target.value)} placeholder="password" type="password" />
          <button type="submit">Login</button>
        </form>
      ) : (
        <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
          <button onClick={loadApps}>İlanları getir</button>
          <button onClick={logout}>Çıkış</button>
        </div>
      )}

      {msg && <p style={{ marginTop: 12 }}>{msg}</p>}

      {apps.length > 0 && (
        <ul style={{ marginTop: 12 }}>
          {apps.map((a) => (
            <li key={a.id}>
              <strong>#{a.id}</strong> — {a.title} @ {a.company} ({a.location})
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
