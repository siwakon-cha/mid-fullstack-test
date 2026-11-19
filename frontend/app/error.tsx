"use client";

export default function ErrorPage({ reset }: { error: Error & { digest?: string }; reset: () => void }) {
  return (
    <main className="shell page">
      <div className="empty"><h2>ไม่สามารถโหลดข้อมูลได้</h2><p>กรุณาตรวจสอบการเชื่อมต่อแล้วลองอีกครั้ง</p><button className="btn" onClick={reset}>ลองใหม่</button></div>
    </main>
  );
}
