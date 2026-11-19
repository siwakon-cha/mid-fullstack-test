"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import type { User } from "@/lib/types";

export default function Nav() {
  const router = useRouter();
  const pathname = usePathname();
  const [user, setUser] = useState<User | null>(null);

  useEffect(() => {
    fetch("/api/auth/me", { cache: "no-store" })
      .then((response) => (response.ok ? response.json() : null))
      .then((data: { user: User } | null) => setUser(data?.user ?? null))
      .catch(() => setUser(null));
  }, [pathname]);

  async function logout() {
    await fetch("/api/auth/logout", { method: "POST" });
    setUser(null);
    router.push("/events");
    router.refresh();
  }

  return (
    <header className="nav">
      <div className="shell nav-inner">
        <Link className="brand" href="/events">จองตั๋ว</Link>
        <nav className="nav-links" aria-label="เมนูหลัก">
          <Link href="/events">อีเวนต์</Link>
          <Link href="/bookings">การจองของฉัน</Link>
          {user?.role === 1 && <Link href="/admin/events">จัดการอีเวนต์</Link>}
          {user ? (
            <button className="btn secondary" type="button" onClick={logout}>ออกจากระบบ</button>
          ) : (
            <><Link href="/login">เข้าสู่ระบบ</Link><Link className="btn" href="/register">สมัครสมาชิก</Link></>
          )}
        </nav>
      </div>
    </header>
  );
}
