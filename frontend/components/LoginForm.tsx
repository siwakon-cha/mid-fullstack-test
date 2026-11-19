"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState, type FormEvent } from "react";
import type { APIError } from "@/lib/types";

export default function LoginForm() {
	const router = useRouter();
  const [error, setError] = useState("");
  const [pending, setPending] = useState(false);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setPending(true);
    setError("");
    const form = new FormData(event.currentTarget);
    const response = await fetch("/api/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email: form.get("email"), password: form.get("password") }),
    });
    if (!response.ok) {
      const data = (await response.json()) as APIError;
      setError(data.error.message);
      setPending(false);
      return;
    }
    router.push("/events");
    router.refresh();
  }

  return (
    <form className="card form-card stack" onSubmit={submit}>
      <div><span className="eyebrow" style={{ color: "var(--brand)" }}>Welcome back</span><h2>เข้าสู่ระบบ</h2></div>
      {error && <p className="message error" role="alert">{error}</p>}
      <div className="field"><label htmlFor="email">อีเมล</label><input className="input" id="email" name="email" type="email" autoComplete="email" required /></div>
      <div className="field"><label htmlFor="password">รหัสผ่าน</label><input className="input" id="password" name="password" type="password" autoComplete="current-password" required /></div>
      <button className="btn" disabled={pending}>{pending ? "กำลังเข้าสู่ระบบ…" : "เข้าสู่ระบบ"}</button>
      <p className="meta">ยังไม่มีบัญชี? <Link href="/register" style={{ color: "var(--brand)", fontWeight: 800 }}>สมัครสมาชิก</Link></p>
    </form>
  );
}
