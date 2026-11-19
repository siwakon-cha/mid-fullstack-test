"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState, type FormEvent } from "react";
import type { APIError } from "@/lib/types";

export default function RegisterForm() {
	const router = useRouter();
  const [error, setError] = useState("");
  const [pending, setPending] = useState(false);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    const form = new FormData(event.currentTarget);
    if (form.get("password") !== form.get("confirm_password")) {
      setError("รหัสผ่านและการยืนยันรหัสผ่านไม่ตรงกัน");
      return;
    }
    setPending(true);
    const response = await fetch("/api/auth/register", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        first_name: form.get("first_name"), last_name: form.get("last_name"), email: form.get("email"),
        password: form.get("password"), phone: form.get("phone") || null,
      }),
    });
    if (!response.ok) {
      const data = (await response.json()) as APIError;
      setError(data.error.message);
      setPending(false);
      return;
    }
    router.push("/login?registered=1");
  }

  return (
    <form className="card form-card stack" onSubmit={submit}>
      <div><span className="eyebrow" style={{ color: "var(--brand)" }}>Create account</span><h2>สมัครสมาชิก</h2></div>
      {error && <p className="message error" role="alert">{error}</p>}
      <div className="form-grid">
        <div className="field"><label htmlFor="first_name">ชื่อจริง</label><input className="input" id="first_name" name="first_name" maxLength={100} required /></div>
        <div className="field"><label htmlFor="last_name">นามสกุล</label><input className="input" id="last_name" name="last_name" maxLength={100} required /></div>
      </div>
      <div className="field"><label htmlFor="email">อีเมล</label><input className="input" id="email" name="email" type="email" autoComplete="email" required /></div>
      <div className="field"><label htmlFor="phone">เบอร์โทร (ไม่บังคับ)</label><input className="input" id="phone" name="phone" type="tel" maxLength={20} /></div>
      <div className="form-grid">
        <div className="field"><label htmlFor="password">รหัสผ่าน</label><input className="input" id="password" name="password" type="password" minLength={8} maxLength={72} autoComplete="new-password" required /></div>
        <div className="field"><label htmlFor="confirm_password">ยืนยันรหัสผ่าน</label><input className="input" id="confirm_password" name="confirm_password" type="password" minLength={8} maxLength={72} autoComplete="new-password" required /></div>
      </div>
      <button className="btn" disabled={pending}>{pending ? "กำลังสร้างบัญชี…" : "สร้างบัญชี"}</button>
      <p className="meta">มีบัญชีแล้ว? <Link href="/login" style={{ color: "var(--brand)", fontWeight: 800 }}>เข้าสู่ระบบ</Link></p>
    </form>
  );
}
