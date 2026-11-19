import type { NextRequest } from "next/server";

export const dynamic = "force-dynamic";

type Context = { params: Promise<{ path: string[] }> };

async function proxy(request: NextRequest, context: Context): Promise<Response> {
  const { path } = await context.params;
  const base = process.env.API_INTERNAL_URL ?? "http://localhost:8080";
  const target = new URL(`/api/${path.join("/")}`, base);
  target.search = request.nextUrl.search;

  const headers = new Headers();
  const contentType = request.headers.get("content-type");
  const cookie = request.headers.get("cookie");
  if (contentType) headers.set("content-type", contentType);
  if (cookie) headers.set("cookie", cookie);
  headers.set("origin", request.headers.get("origin") ?? process.env.FRONTEND_ORIGIN ?? request.nextUrl.origin);

  const hasBody = !["GET", "HEAD"].includes(request.method);
  const upstream = await fetch(target, {
    method: request.method,
    headers,
    body: hasBody ? await request.arrayBuffer() : undefined,
    cache: "no-store",
    redirect: "manual",
  });
  const responseHeaders = new Headers();
  const upstreamType = upstream.headers.get("content-type");
  const setCookie = upstream.headers.get("set-cookie");
  const cacheStatus = upstream.headers.get("x-cache");
  if (upstreamType) responseHeaders.set("content-type", upstreamType);
  if (setCookie) responseHeaders.set("set-cookie", setCookie);
  if (cacheStatus) responseHeaders.set("x-cache", cacheStatus);
  return new Response(upstream.body, { status: upstream.status, headers: responseHeaders });
}

export const GET = proxy;
export const POST = proxy;
export const PATCH = proxy;
export const DELETE = proxy;
export const PUT = proxy;
