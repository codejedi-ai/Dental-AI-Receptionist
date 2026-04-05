import { NextRequest, NextResponse } from "next/server";
import { getAvailableSlots } from "@/lib/db";

export async function GET(req: NextRequest) {
  const url = new URL(req.url);
  const date = url.searchParams.get("date");
  const dentist = url.searchParams.get("dentist");

  if (!date || !dentist) {
    return NextResponse.json({ error: "date and dentist are required" }, { status: 400 });
  }

  const slots = getAvailableSlots(date, dentist);
  return NextResponse.json({ slots, date, dentist });
}
